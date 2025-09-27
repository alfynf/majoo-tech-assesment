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
	ErrPostNotFound = errors.New("post not found")
	ErrForbidden    = errors.New("user is not allowed to perform this action")
)

// PostService mendefinisikan logika bisnis untuk Post
type PostService interface {
	CreatePost(ctx context.Context, userID uuid.UUID, title, content string) (*model.Post, error)
	GetPostByID(ctx context.Context, postID uuid.UUID) (*model.Post, error)
	GetPostsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Post, error)
	UpdatePost(ctx context.Context, postID, userID uuid.UUID, title, content string) (*model.Post, error)
	DeletePost(ctx context.Context, postID, userID uuid.UUID) error
}

type postService struct {
	postRepo repository.PostRepository
}

// NewPostService adalah constructor untuk postService
func NewPostService(repo repository.PostRepository) PostService {
	return &postService{
		postRepo: repo,
	}
}

// CreatePost membuat post baru
func (s *postService) CreatePost(ctx context.Context, userID uuid.UUID, title, content string) (*model.Post, error) {
	post := &model.Post{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   title,
		Content: content,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Mengambil post yang baru dibuat untuk mendapatkan timestamp dari database
	createdPost, err := s.postRepo.FindByID(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created post: %w", err)
	}

	return createdPost, nil
}

// GetPostByID mengambil satu post berdasarkan ID
func (s *postService) GetPostByID(ctx context.Context, postID uuid.UUID) (*model.Post, error) {
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	return post, nil
}

// GetPostsByUserID mengambil semua post dari seorang user
func (s *postService) GetPostsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Post, error) {
	return s.postRepo.FindByUserID(ctx, userID)
}

// UpdatePost memperbarui post yang ada
func (s *postService) UpdatePost(ctx context.Context, postID, userID uuid.UUID, title, content string) (*model.Post, error) {
	post, err := s.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Otorisasi: Pastikan user yang mengedit adalah pemilik post
	if post.UserID != userID {
		return nil, ErrForbidden
	}

	post.Title = title
	post.Content = content

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Mengambil post yang sudah diupdate untuk mendapatkan timestamp baru
	return s.GetPostByID(ctx, postID)
}

// DeletePost menghapus sebuah post
func (s *postService) DeletePost(ctx context.Context, postID, userID uuid.UUID) error {
	post, err := s.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}

	// Otorisasi: Pastikan user yang menghapus adalah pemilik post
	if post.UserID != userID {
		return ErrForbidden
	}

	return s.postRepo.Delete(ctx, postID)
}
