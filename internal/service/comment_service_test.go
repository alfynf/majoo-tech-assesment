package service_test

import (
	"context"
	"database/sql"
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

func TestCommentService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
	mockPostRepo := mocks.NewMockPostRepository(ctrl) // Comment service depends on Post repo too
	commentService := service.NewCommentService(mockCommentRepo, mockPostRepo)
	ctx := context.Background()

	userID := uuid.New()
	postID := uuid.New()
	commentID := uuid.New()

	post := &model.Post{ID: postID, UserID: uuid.New()}
	comment := &model.Comment{
		ID:        commentID,
		PostID:    postID,
		UserID:    userID,
		Content:   "Test Comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("CreateComment - Success", func(t *testing.T) {
		// 1. Check if post exists
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(post, nil)
		// 2. Create the comment
		mockCommentRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		// 3. Fetch the created comment
		mockCommentRepo.EXPECT().FindByID(ctx, gomock.Any()).Return(comment, nil)

		createdComment, err := commentService.CreateComment(ctx, userID, postID, "Test Comment")
		assert.NoError(t, err)
		assert.Equal(t, commentID, createdComment.ID)
	})

	t.Run("CreateComment - Post Not Found", func(t *testing.T) {
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(nil, nil)

		_, err := commentService.CreateComment(ctx, userID, postID, "Test Comment")
		assert.ErrorIs(t, err, service.ErrPostNotFound)
	})

	t.Run("CreateComment - Find Post Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(nil, dbErr)

		_, err := commentService.CreateComment(ctx, userID, postID, "Test Comment")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find post for comment")
	})

	t.Run("CreateComment - Create Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockPostRepo.EXPECT().FindByID(ctx, postID).Return(post, nil)
		mockCommentRepo.EXPECT().Create(ctx, gomock.Any()).Return(dbErr)

		_, err := commentService.CreateComment(ctx, userID, postID, "Test Comment")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create comment")
	})

	t.Run("DeleteComment - Success", func(t *testing.T) {
		// 1. Find the comment to check ownership
		mockCommentRepo.EXPECT().FindByID(ctx, commentID).Return(comment, nil)
		// 2. Delete the comment
		mockCommentRepo.EXPECT().Delete(ctx, commentID).Return(nil)

		err := commentService.DeleteComment(ctx, commentID, userID)
		assert.NoError(t, err)
	})

	t.Run("DeleteComment - Forbidden", func(t *testing.T) {
		wrongUserID := uuid.New()
		mockCommentRepo.EXPECT().FindByID(ctx, commentID).Return(comment, nil)

		err := commentService.DeleteComment(ctx, commentID, wrongUserID)
		assert.ErrorIs(t, err, service.ErrForbidden)
	})

	t.Run("DeleteComment - Not Found", func(t *testing.T) {
		mockCommentRepo.EXPECT().FindByID(ctx, commentID).Return(nil, sql.ErrNoRows)

		err := commentService.DeleteComment(ctx, commentID, userID)
		assert.ErrorIs(t, err, service.ErrCommentNotFound)
	})

	t.Run("UpdateComment - Success", func(t *testing.T) {
		updatedContent := "Updated Content"
		mockCommentRepo.EXPECT().FindByID(ctx, commentID).Return(comment, nil).Times(2)
		mockCommentRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, c *model.Comment) error {
			assert.Equal(t, updatedContent, c.Content)
			return nil
		})

		updatedComment, err := commentService.UpdateComment(ctx, commentID, userID, updatedContent)
		assert.NoError(t, err)
		assert.NotNil(t, updatedComment)
	})

	t.Run("UpdateComment - Not Found on Find", func(t *testing.T) {
		mockCommentRepo.EXPECT().FindByID(ctx, commentID).Return(nil, sql.ErrNoRows)

		_, err := commentService.UpdateComment(ctx, commentID, userID, "new content")
		assert.ErrorIs(t, err, service.ErrCommentNotFound)
	})

	t.Run("UpdateComment - Find Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockCommentRepo.EXPECT().FindByID(ctx, commentID).Return(nil, dbErr)

		_, err := commentService.UpdateComment(ctx, commentID, userID, "new content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find comment for update")
	})

	t.Run("UpdateComment - Forbidden", func(t *testing.T) {
		wrongUserID := uuid.New()
		mockCommentRepo.EXPECT().FindByID(ctx, commentID).Return(comment, nil)

		_, err := commentService.UpdateComment(ctx, commentID, wrongUserID, "new content")
		assert.ErrorIs(t, err, service.ErrForbidden)
	})

	t.Run("UpdateComment - Update Fails", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockCommentRepo.EXPECT().FindByID(ctx, commentID).Return(comment, nil)
		mockCommentRepo.EXPECT().Update(ctx, gomock.Any()).Return(dbErr)

		_, err := commentService.UpdateComment(ctx, commentID, userID, "new content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update comment")
	})
}
