package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"colorpixel/internal/timeutil"
)

var ErrNotFound = errors.New("not found")

func (d *DB) EnsureUser(ctx context.Context, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = d.Pool.Exec(ctx, `
INSERT INTO users (username, password_hash, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (username) DO NOTHING`, username, string(hash), timeutil.Now())
	return err
}

func (d *DB) Authenticate(ctx context.Context, username, password string) (*User, error) {
	var u User
	err := d.Pool.QueryRow(ctx, `SELECT id, username, password_hash FROM users WHERE username=$1`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, ErrNotFound
	}
	return &u, nil
}
