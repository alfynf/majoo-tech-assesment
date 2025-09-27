package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrDuplicateEntry = errors.New("violates unique constraint")
)

// UserRepository defines database operations for a User.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// userRepository adalah implementasi konkret dari UserRepository
type userRepository struct {
	db *sql.DB
}

// NewUserRepository adalah constructor untuk userRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Create menyimpan user baru ke database
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, username, email, password_hash) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.Password)
	if err != nil {
		// Check if the error is a unique violation from postgres.
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateEntry
		}
		return err
	}
	return nil
}

// FindByEmail mencari user berdasarkan alamat email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

// FindByID mencari user berdasarkan ID
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}
