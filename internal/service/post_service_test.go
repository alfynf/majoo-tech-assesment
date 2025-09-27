package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/alfynf/simple-blog-api/internal/repository/mocks"
	"github.com/alfynf/simple-blog-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostRepo := mocks.NewMockPostRepository(ctrl)
	postService := service.NewPostService(mockPostRepo)
	ctx := context.Background()

	userID := uuid.New()
	postID := uuid.New()
	post := &model.Post{
		ID:        postID,
		UserID:    userID,
		Title:     "Test Title",
		Content:   "Test Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("CreatePost - Success", func(t *testing.T) {
		mockPostRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		// After creating, the service fetches the post again to get DB-generated values
		mockPostRepo.EXPECT().FindByID(ctx, gomock.Any()).Return(post, nil)

		createdPost, err := postService.CreatePost(ctx, userID, "Test Title", "Test Content")
		assert.NoError(t, err)
		assert.Equal(t, postID, createdPost.ID)
	})

	t.Run("CreatePost - Create Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockPostRepo.EXPECT().Create(ctx, gomock.Any()).Return(dbErr)

		_, err := postService.CreatePost(ctx, userID, "Test Title", "Test Content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create post")
	})

	t.Run("CreatePost - Retrieve Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockPostRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockPostRepo.EXPECT().FindByID(ctx, gomock.Any()).Return(nil, dbErr)

		_, err := postService.CreatePost(ctx, userID, "Test Title", "Test Content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve created post")
	})

	t.Run("GetPostByID - Success", func(t *testing.T) {
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(post, nil)

		foundPost, err := postService.GetPostByID(ctx, postID)
		if err != nil {
			t.Fatalf("GetPostByID() error = %v, wantErr %v", err, false)
		}
		assert.Equal(t, postID, foundPost.ID)
	})

	t.Run("GetPostByID - Not Found", func(t *testing.T) {
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(nil, nil) // Repository returns nil, nil for not found

		_, err := postService.GetPostByID(ctx, postID)
		assert.ErrorIs(t, err, service.ErrPostNotFound)
	})

	t.Run("GetPostByID - DB Error", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(nil, dbErr)

		_, err := postService.GetPostByID(ctx, postID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get post by id")
	})

	t.Run("UpdatePost - Success", func(t *testing.T) {
		updatedTitle := "Updated Title"
		updatedContent := "Updated Content"

		// First, the service finds the post
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(post, nil)
		// Then, it updates it
		mockPostRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)
		// Finally, it fetches the updated post
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(&model.Post{
			ID:      postID,
			UserID:  userID,
			Title:   updatedTitle,
			Content: updatedContent,
		}, nil)

		updatedPost, err := postService.UpdatePost(ctx, postID, userID, updatedTitle, updatedContent)
		assert.NoError(t, err)
		assert.Equal(t, updatedTitle, updatedPost.Title)
	})

	t.Run("UpdatePost - Forbidden", func(t *testing.T) {
		wrongUserID := uuid.New()
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(post, nil)

		_, err := postService.UpdatePost(ctx, postID, wrongUserID, "title", "content")
		assert.ErrorIs(t, err, service.ErrForbidden)
	})

	t.Run("UpdatePost - GetPostByID Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(nil, dbErr)

		_, err := postService.UpdatePost(ctx, postID, userID, "title", "content")
		assert.Error(t, err)
	})

	t.Run("UpdatePost - Update Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(post, nil)
		mockPostRepo.EXPECT().Update(ctx, gomock.Any()).Return(dbErr)

		_, err := postService.UpdatePost(ctx, postID, userID, "title", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update post")
	})

	t.Run("DeletePost - Success", func(t *testing.T) {
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(post, nil)
		mockPostRepo.EXPECT().Delete(ctx, postID).Return(nil)

		err := postService.DeletePost(ctx, postID, userID)
		assert.NoError(t, err)
	})

	t.Run("DeletePost - Forbidden", func(t *testing.T) {
		wrongUserID := uuid.New()
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(post, nil)

		err := postService.DeletePost(ctx, postID, wrongUserID)
		assert.ErrorIs(t, err, service.ErrForbidden)
	})

	t.Run("DeletePost - GetPostByID Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(nil, dbErr)

		err := postService.DeletePost(ctx, postID, userID)
		assert.Error(t, err)
	})
}
