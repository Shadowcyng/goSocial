package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id,omitempty"`
	Role      Role     `json:"role,omitempty"`
}
type Password struct {
	text *string
	hash []byte
}

func (p *Password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.hash = hash
	p.text = &text
	return nil
}

func (p *Password) Validate(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) GetById(ctx context.Context, userId int64) (*User, error) {
	query := `SELECT 
    users.id AS user_id, 
    username, 
    email, 
    password, 
    created_at, 
    is_active, 
    roles.* 
	FROM users
	JOIN roles ON roles.id = users.role_id
	WHERE users.id = $1 AND is_active = true;`
	user, err := s.getUser(ctx, query, &userId, "")
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `Select id, username, email, password, created_at, is_active FROM users where email = $1 and is_active= true`
	user, err := s.getUser(ctx, query, nil, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `Select id, username, email, password, created_at, is_active FROM users where username = $1 and is_active= true`
	user, err := s.getUser(ctx, query, nil, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.deleteUserInvitations(ctx, tx, userID); err != nil {
			return err
		}

		if err := s.deleteUserById(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	// transaction wrapper
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// create user
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}
		// create invite
		err := s.createUserInviations(ctx, tx, invitationExp, token, user.ID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// find the user that this token belongs to

		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}
		// update active status of user
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}
		// clear invitation
		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

// should be private

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `INSERT into users (username, password, email, role_id) 
	VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) returning id, created_at
	`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	var role string
	if user.Role.Name == "" {
		role = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password.hash,
		user.Email,
		role,
	).Scan(
		&user.ID,
		&user.CreatedAt)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrorDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrorDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

// private

func (s *UserStore) createUserInviations(ctx context.Context, tx *sql.Tx, exp time.Duration, token string, userID int64) error {
	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `SELECT u.id, u.username, u.email, u.created_at, u.is_active 
	FROM users u
	LEFT JOIN user_invitations ui ON u.id = ui.user_id
	WHERE ui.token = $1 AND ui.expiry > $2
	`
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString((hash[:]))

	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()
	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active = $3 where id = $4`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations where user_id = $1`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) deleteUserById(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM users where id = $1`
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) getUser(ctx context.Context, query string, id *int64, param string) (*User, error) {
	ctx, cancelCtx := context.WithTimeout(ctx, QueryTimeout)
	defer cancelCtx()
	var user User
	var err error
	if id != nil {
		err = s.db.QueryRowContext(ctx, query, *id).Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.CreatedAt,
			&user.IsActive,
			&user.Role.ID,
			&user.Role.Name,
			&user.Role.Level,
			&user.Role.Description,
		)
	} else {
		err = s.db.QueryRowContext(ctx, query, param).Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.CreatedAt,
			&user.IsActive,
		)
	}
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}
