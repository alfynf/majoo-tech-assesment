package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alfynf/simple-blog-api/internal/handler"
	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/alfynf/simple-blog-api/internal/service"
	"github.com/alfynf/simple-blog-api/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCommentHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	commentHandler := handler.NewCommentHandler(mockCommentService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	userID := uuid.New()
	setUserID := func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}

	// Public routes
	router.GET("/posts/:id/comments", commentHandler.GetCommentsByPost)

	// Protected routes
	protected := router.Group("/")
	protected.Use(setUserID)
	{
		protected.POST("/posts/:id/comments", commentHandler.CreateComment)
		protected.PUT("/comments/:commentID", commentHandler.UpdateComment)
		protected.DELETE("/comments/:commentID", commentHandler.DeleteComment)
	}

	t.Run("CreateComment - Success", func(t *testing.T) {
		postID := uuid.New()
		reqBody := handler.CreateCommentRequest{Content: "A new comment"}
		comment := &model.Comment{ID: uuid.New(), PostID: postID, UserID: userID, Content: reqBody.Content}

		mockCommentService.EXPECT().CreateComment(gomock.Any(), userID, postID, reqBody.Content).Return(comment, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%s/comments", postID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var respBody model.Comment
		json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.Equal(t, comment.ID, respBody.ID)
	})

	t.Run("CreateComment - Post Not Found", func(t *testing.T) {
		postID := uuid.New()
		reqBody := handler.CreateCommentRequest{Content: "A new comment"}

		mockCommentService.EXPECT().CreateComment(gomock.Any(), userID, postID, reqBody.Content).Return(nil, service.ErrPostNotFound)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%s/comments", postID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("CreateComment - Invalid Post ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/posts/not-a-uuid/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateComment - Binding Failure", func(t *testing.T) {
		postID := uuid.New()
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%s/comments", postID), bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateComment - Service Failure", func(t *testing.T) {
		postID := uuid.New()
		reqBody := handler.CreateCommentRequest{Content: "A new comment"}
		mockCommentService.EXPECT().CreateComment(gomock.Any(), userID, postID, reqBody.Content).Return(nil, errors.New("internal error"))

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%s/comments", postID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("CreateComment - User Not Authenticated", func(t *testing.T) {
		unauthedRouter := gin.Default()
		unauthedRouter.POST("/posts/:id/comments", commentHandler.CreateComment)
		postID := uuid.New()
		reqBody := handler.CreateCommentRequest{Content: "A new comment"}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%s/comments", postID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		unauthedRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GetCommentsByPost - Success", func(t *testing.T) {
		postID := uuid.New()
		comments := []model.Comment{{ID: uuid.New()}, {ID: uuid.New()}}

		mockCommentService.EXPECT().GetCommentsByPostID(gomock.Any(), postID).Return(comments, nil)

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s/comments", postID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respBody []model.Comment
		json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.Len(t, respBody, 2)
	})

	t.Run("GetCommentsByPost - Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/posts/not-a-uuid/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetCommentsByPost - Service Failure", func(t *testing.T) {
		postID := uuid.New()
		mockCommentService.EXPECT().GetCommentsByPostID(gomock.Any(), postID).Return(nil, errors.New("internal error"))

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s/comments", postID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to retrieve comments")
	})

	t.Run("UpdateComment - Forbidden", func(t *testing.T) {
		commentID := uuid.New()
		reqBody := handler.UpdateCommentRequest{Content: "Updated content"}

		mockCommentService.EXPECT().UpdateComment(gomock.Any(), commentID, userID, reqBody.Content).Return(nil, service.ErrForbidden)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s", commentID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("UpdateComment - Not Found", func(t *testing.T) {
		commentID := uuid.New()
		reqBody := handler.UpdateCommentRequest{Content: "Updated content"}

		mockCommentService.EXPECT().UpdateComment(gomock.Any(), commentID, userID, reqBody.Content).Return(nil, service.ErrCommentNotFound)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s", commentID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("UpdateComment - Success", func(t *testing.T) {
		commentID := uuid.New()
		reqBody := handler.UpdateCommentRequest{Content: "Updated content"}
		updatedComment := &model.Comment{ID: commentID, Content: reqBody.Content}

		mockCommentService.EXPECT().UpdateComment(gomock.Any(), commentID, userID, reqBody.Content).Return(updatedComment, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s", commentID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respBody model.Comment
		json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.Equal(t, updatedComment.Content, respBody.Content)
	})

	t.Run("UpdateComment - Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/comments/not-a-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateComment - Binding Failure", func(t *testing.T) {
		commentID := uuid.New()
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s", commentID), bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateComment - Service Failure", func(t *testing.T) {
		commentID := uuid.New()
		reqBody := handler.UpdateCommentRequest{Content: "Updated content"}
		mockCommentService.EXPECT().UpdateComment(gomock.Any(), commentID, userID, reqBody.Content).Return(nil, errors.New("internal error"))

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s", commentID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UpdateComment - User Not Authenticated", func(t *testing.T) {
		unauthedRouter := gin.Default()
		unauthedRouter.PUT("/comments/:commentID", commentHandler.UpdateComment)

		// Provide a valid body so the binding doesn't fail before the auth check.
		reqBody := handler.UpdateCommentRequest{Content: "some valid content"}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s", uuid.New()), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		unauthedRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("DeleteComment - Success", func(t *testing.T) {
		commentID := uuid.New()
		mockCommentService.EXPECT().DeleteComment(gomock.Any(), commentID, userID).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/comments/%s", commentID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("DeleteComment - Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/comments/not-a-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DeleteComment - Service Failure", func(t *testing.T) {
		commentID := uuid.New()
		mockCommentService.EXPECT().DeleteComment(gomock.Any(), commentID, userID).Return(errors.New("internal error"))
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/comments/%s", commentID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("DeleteComment - User Not Authenticated", func(t *testing.T) {
		unauthedRouter := gin.Default()
		unauthedRouter.DELETE("/comments/:commentID", commentHandler.DeleteComment)
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/comments/%s", uuid.New()), nil)
		w := httptest.NewRecorder()
		unauthedRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
