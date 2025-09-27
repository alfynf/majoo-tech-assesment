package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/google/uuid"
)

// CommentRepository mendefinisikan operasi database untuk Comment
type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	FindByPostID(ctx context.Context, postID uuid.UUID) ([]model.Comment, error)
	Update(ctx context.Context, comment *model.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// commentRepository adalah implementasi konkret dari CommentRepository
type commentRepository struct {
	db *sql.DB
}

// NewCommentRepository adalah constructor untuk commentRepository
func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

// Create menyimpan comment baru ke database
func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	query := `INSERT INTO comments (id, post_id, user_id, content) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, comment.ID, comment.PostID, comment.UserID, comment.Content)
	return err
}

// FindByID mencari comment berdasarkan ID
func (r *commentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	comment := &model.Comment{}
	query := `SELECT id, post_id, user_id, content, created_at, updated_at FROM comments WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Comment tidak ditemukan
	}
	return comment, err
}

// FindByPostID mencari semua comment berdasarkan PostID
func (r *commentRepository) FindByPostID(ctx context.Context, postID uuid.UUID) ([]model.Comment, error) {
	query := `SELECT id, post_id, user_id, content, created_at, updated_at FROM comments WHERE post_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil comments berdasarkan post ID: %w", err)
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var comment model.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai baris comment: %w", err)
		}
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error setelah iterasi baris: %w", err)
	}
	return comments, nil
}

// Update memperbarui comment yang sudah ada di database
func (r *commentRepository) Update(ctx context.Context, comment *model.Comment) error {
	query := `UPDATE comments SET content = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, comment.Content, comment.ID)
	return err
}

// Delete menghapus comment dari database
func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
