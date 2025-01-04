package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/Shadowcyng/goSocial/internal/store"
)

func Seed(store store.Storage, db *sql.DB) error {
	ctx := context.Background()
	users := generateUser(1000)
	tx, _ := db.BeginTx(ctx, nil)
	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			log.Println("Error creating user", err)
			return err
		}
	}
	tx.Commit()
	posts := generatePost(100000, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			_ = tx.Rollback()
			log.Println("Error creating post", err)
			return err
		}
	}

	comments := generateComments(1000000, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("Error creating comment", err)
			return err
		}
	}
	log.Println("Seeding completed successfully")
	return nil
}

func generateUser(num int) []*store.User {
	password := store.Password{}
	users := make([]*store.User, num)
	for i := 0; i < num; i++ {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Password: password,
		}
	}
	return users
}

func generatePost(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)
	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]
		posts[i] = &store.Post{
			Title:   titles[rand.Intn(len(titles))],
			Content: contents[rand.Intn(len(contents))],
			UserID:  user.ID,
			Tags: []string{
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
			},
		}
	}
	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, num)
	for i := 0; i < num; i++ {
		user := *users[rand.Intn(len(users))]
		post := posts[rand.Intn(len(posts))]
		comments[i] = &store.Comment{
			PostID:  post.ID,
			User:    user,
			UserID:  user.ID,
			Content: commentsContent[rand.Intn(len(commentsContent))],
		}
	}
	return comments
}
