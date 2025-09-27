package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/alfynf/simple-blog-api/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
)

// CommentService mendefinisikan logika bisnis untuk Comment
type CommentService interface {
	CreateComment(ctx context.Context, userID, postID uuid.UUID, content string) (*model.Comment, error)
	GetCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]model.Comment, error)
	UpdateComment(ctx context.Context, commentID, userID uuid.UUID, content string) (*model.Comment, error)
	DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository // Untuk memeriksa keberadaan post
}

// NewCommentService adalah constructor untuk commentService
func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

// CreateComment membuat komentar baru pada sebuah post
func (s *commentService) CreateComment(ctx context.Context, userID, postID uuid.UUID, content string) (*model.Comment, error) {
	// 1. Pastikan post ada sebelum menambahkan komentar
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to find post for comment: %w", err)
	}
	if post == nil {
		return nil, ErrPostNotFound // Menggunakan error dari post_service
	}

	// 2. Buat model komentar
	comment := &model.Comment{
		ID:      uuid.New(),
		PostID:  postID,
		UserID:  userID,
		Content: content,
	}

	// 3. Simpan ke database
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// 4. Ambil kembali untuk mendapatkan data lengkap (termasuk timestamp)
	return s.commentRepo.FindByID(ctx, comment.ID)
}

// GetCommentsByPostID mengambil semua komentar dari sebuah post
func (s *commentService) GetCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]model.Comment, error) {
	return s.commentRepo.FindByPostID(ctx, postID)
}

// UpdateComment memperbarui sebuah komentar
func (s *commentService) UpdateComment(ctx context.Context, commentID, userID uuid.UUID, content string) (*model.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to find comment for update: %w", err)
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}

	// Otorisasi: Pastikan user yang mengedit adalah pemilik komentar
	if comment.UserID != userID {
		return nil, ErrForbidden
	}

	comment.Content = content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return s.commentRepo.FindByID(ctx, commentID)
}

// DeleteComment menghapus sebuah komentar
func (s *commentService) DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil || comment == nil {
		return ErrCommentNotFound
	}

	// Otorisasi: Pastikan user yang menghapus adalah pemilik komentar
	if comment.UserID != userID {
		return ErrForbidden
	}

	return s.commentRepo.Delete(ctx, commentID)
}
