package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post           Post     `json:"post"`
	CommentCount   int      `json:"comment_count"`
	LatestComments []string `json:"latest_comments"`
}
type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (content, title, user_id, tags) 
	VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	err := s.db.QueryRowContext(ctx, query, post.Content, post.Title, post.UserID, pq.Array(post.Tags)).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetById(ctx context.Context, postID int64) (*Post, error) {
	query := `Select id, user_id, title, content, created_at, updated_at, tags, version
	 FROM posts 
	 WHERE id = $1`

	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()
	var post Post
	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostStore) DeleteById(ctx context.Context, postID int64) error {
	query := `DELETE FROM posts WHERE id = $1 returning id, user_id, title, content, created_at, updated_at, tags`

	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrorNotFound
	}
	return nil
}

func (s *PostStore) UpdatePostById(ctx context.Context, post *Post) error {
	query := `UPDATE posts
	SET title = $1, content = $2, tags = $3, version = version + 1
	where id = $4 AND version = $5
	RETURNING version;
	`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, pq.Array(post.Tags), post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, ErrorNotFound):
			return ErrorNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, id int64, fq PaginatedFeedQuery) ([]*PostWithMetadata, error) {
	query := fmt.Sprintf(`SELECT 
    p.id, 
    p.user_id, 
    p.title, 
    p.content, 
    p.created_at, 
    p.version, 
    p.tags, 
    u.username,
    COUNT(c.id) AS comments_count,
    COALESCE(
        (SELECT ARRAY_AGG(lc.comment_content ORDER BY lc.comment_created_at DESC) 
         FROM (
             SELECT content AS comment_content, created_at AS comment_created_at 
             FROM comments 
             WHERE post_id = p.id AND content IS NOT NULL
             ORDER BY created_at DESC 
             LIMIT 2
         ) lc),
        '{}'::TEXT[]
    ) AS latest_comments
	FROM posts p
	LEFT JOIN users u ON u.id = p.user_id
	LEFT JOIN followers f ON f.follower_id = p.user_id
	LEFT JOIN comments c ON c.post_id = p.id
	WHERE 
		(f.user_id = $1 OR p.user_id = $1 ) AND
		(p.title ILIKE '%%' || $4 || '%%' OR p.content ILIKE '%%' || $4 || '%%') AND
		(p.tags @> $5 OR $5 IS NULL OR $5 = '{}'::VARCHAR[])
	GROUP BY p.id, u.username
	ORDER BY %s %s
	LIMIT $2 OFFSET $3;
	`, fq.SortBy, fq.SortOrder)

	fmt.Println(query)
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()
	rows, err := s.db.QueryContext(ctx, query, id, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err != nil {
		return nil, err
	}
	var feed []*PostWithMetadata
	for rows.Next() {
		var post PostWithMetadata
		err := rows.Scan(
			&post.Post.ID,
			&post.Post.UserID,
			&post.Post.Title,
			&post.Post.Content,
			&post.Post.CreatedAt,
			&post.Post.Version,
			pq.Array(&post.Post.Tags),
			&post.Post.User.Username,
			&post.CommentCount,
			pq.Array(&post.LatestComments),
		)
		if err != nil {
			return nil, err
		}
		feed = append(feed, &post)
	}
	return feed, nil
}
