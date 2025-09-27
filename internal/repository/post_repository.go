package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/google/uuid"
)

// PostRepository mendefinisikan operasi database untuk Post
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Post, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Post, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// postRepository adalah implementasi konkret dari PostRepository
type postRepository struct {
	db *sql.DB
}

// NewPostRepository adalah constructor untuk postRepository
func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

// Create menyimpan post baru ke database
func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	query := `INSERT INTO posts (id, user_id, title, content) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, post.ID, post.UserID, post.Title, post.Content)
	return err
}

// FindByID mencari post berdasarkan ID
func (r *postRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	post := &model.Post{}
	query := `SELECT id, user_id, title, content, created_at, updated_at FROM posts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Post tidak ditemukan
	}
	return post, err
}

// FindByUserID mencari semua post berdasarkan UserID
func (r *postRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Post, error) {
	query := `SELECT id, user_id, title, content, created_at, updated_at FROM posts WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil posts berdasarkan user ID: %w", err)
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai baris post: %w", err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error setelah iterasi baris: %w", err)
	}
	return posts, nil
}

// Update memperbarui post yang sudah ada di database
func (r *postRepository) Update(ctx context.Context, post *model.Post) error {
	query := `UPDATE posts SET title = $1, content = $2, updated_at = NOW() WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, post.ID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan jumlah baris yang terpengaruh: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("post dengan ID %s tidak ditemukan untuk diperbarui", post.ID)
	}
	return nil
}

// Delete menghapus post dari database
func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus post: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan jumlah baris yang terpengaruh: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("post dengan ID %s tidak ditemukan untuk dihapus", id)
	}
	return nil
}
