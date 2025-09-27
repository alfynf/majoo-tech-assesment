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

func TestPostHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostService := mocks.NewMockPostService(ctrl)
	postHandler := handler.NewPostHandler(mockPostService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Helper to set user ID in context for authenticated routes
	setUserID := func(id uuid.UUID) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("userID", id)
			c.Next()
		}
	}

	userID := uuid.New()

	// Public routes
	router.GET("/posts/:id", postHandler.GetPost)

	// Protected routes
	protected := router.Group("/")
	protected.Use(setUserID(userID))
	{
		protected.POST("/posts", postHandler.CreatePost)
		protected.PUT("/posts/:id", postHandler.UpdatePost)
		protected.DELETE("/posts/:id", postHandler.DeletePost)
	}

	t.Run("CreatePost - Success", func(t *testing.T) {
		reqBody := handler.CreatePostRequest{Title: "New Post", Content: "Some content"}
		post := &model.Post{ID: uuid.New(), UserID: userID, Title: reqBody.Title, Content: reqBody.Content}

		mockPostService.EXPECT().CreatePost(gomock.Any(), userID, reqBody.Title, reqBody.Content).Return(post, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var respBody model.Post
		json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.Equal(t, post.ID, respBody.ID)
	})

	t.Run("CreatePost - Binding Failure", func(t *testing.T) {
		reqBody := map[string]string{"title": "t"} // Invalid content
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreatePost - User Not Authenticated", func(t *testing.T) {
		// This test uses a router without the setUserID middleware
		unauthedRouter := gin.Default()
		unauthedRouter.POST("/posts", postHandler.CreatePost)

		reqBody := handler.CreatePostRequest{Title: "New Post", Content: "Some content"}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		unauthedRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("CreatePost - Service Failure", func(t *testing.T) {
		mockPostService.EXPECT().CreatePost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("internal error"))
		req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBufferString(`{"title":"a valid title","content":"some valid content"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetPost - Success", func(t *testing.T) {
		postID := uuid.New()
		post := &model.Post{ID: postID}
		mockPostService.EXPECT().GetPostByID(gomock.Any(), postID).Return(post, nil)

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", postID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respBody model.Post
		json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.Equal(t, postID, respBody.ID)
	})

	t.Run("GetPost - Not Found", func(t *testing.T) {
		postID := uuid.New()
		mockPostService.EXPECT().GetPostByID(gomock.Any(), postID).Return(nil, service.ErrPostNotFound)

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", postID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GetPost - Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/posts/not-a-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid post ID format")
	})

	t.Run("GetPost - Service Failure", func(t *testing.T) {
		postID := uuid.New()
		mockPostService.EXPECT().GetPostByID(gomock.Any(), postID).Return(nil, errors.New("internal error"))

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", postID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to retrieve post")
	})

	t.Run("UpdatePost - Success", func(t *testing.T) {
		postID := uuid.New()
		reqBody := handler.UpdatePostRequest{Title: "Updated Title", Content: "Updated content"}
		updatedPost := &model.Post{ID: postID, UserID: userID, Title: reqBody.Title, Content: reqBody.Content}

		mockPostService.EXPECT().UpdatePost(gomock.Any(), postID, userID, reqBody.Title, reqBody.Content).Return(updatedPost, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%s", postID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respBody model.Post
		json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.Equal(t, reqBody.Title, respBody.Title)
	})

	t.Run("UpdatePost - Forbidden", func(t *testing.T) {
		postID := uuid.New()
		reqBody := handler.UpdatePostRequest{Title: "Updated Title", Content: "Updated content"}

		mockPostService.EXPECT().UpdatePost(gomock.Any(), postID, userID, reqBody.Title, reqBody.Content).Return(nil, service.ErrForbidden)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%s", postID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("UpdatePost - Not Found", func(t *testing.T) {
		postID := uuid.New()
		reqBody := handler.UpdatePostRequest{Title: "Updated Title", Content: "Updated content"}
		mockPostService.EXPECT().UpdatePost(gomock.Any(), postID, userID, reqBody.Title, reqBody.Content).Return(nil, service.ErrPostNotFound)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%s", postID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("UpdatePost - Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/posts/not-a-uuid", bytes.NewBufferString(`{}`))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdatePost - Binding Failure", func(t *testing.T) {
		postID := uuid.New()
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%s", postID), bytes.NewBufferString(`{"title":"t"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdatePost - Service Failure", func(t *testing.T) {
		postID := uuid.New()
		reqBody := handler.UpdatePostRequest{Title: "Updated Title", Content: "Updated content"}
		mockPostService.EXPECT().UpdatePost(gomock.Any(), postID, userID, reqBody.Title, reqBody.Content).Return(nil, errors.New("internal error"))

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%s", postID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UpdatePost - User Not Authenticated", func(t *testing.T) {
		unauthedRouter := gin.Default()
		unauthedRouter.PUT("/posts/:id", postHandler.UpdatePost)

		// Provide a valid body so the binding doesn't fail before the auth check.
		reqBody := handler.UpdatePostRequest{Title: "A valid title", Content: "Some valid content"}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%s", uuid.New()), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		unauthedRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("DeletePost - Success", func(t *testing.T) {
		postID := uuid.New()
		mockPostService.EXPECT().DeletePost(gomock.Any(), postID, userID).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%s", postID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("DeletePost - Not Found or Forbidden", func(t *testing.T) {
		postID := uuid.New()
		// Testing the combined error response for security
		mockPostService.EXPECT().DeletePost(gomock.Any(), postID, userID).Return(service.ErrForbidden)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%s", postID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Post not found or not owned by user")
	})

	t.Run("DeletePost - Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/posts/not-a-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DeletePost - Service Failure", func(t *testing.T) {
		postID := uuid.New()
		mockPostService.EXPECT().DeletePost(gomock.Any(), postID, userID).Return(errors.New("internal error"))
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%s", postID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("DeletePost - User Not Authenticated", func(t *testing.T) {
		unauthedRouter := gin.Default()
		unauthedRouter.DELETE("/posts/:id", postHandler.DeletePost)
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%s", uuid.New()), nil)
		w := httptest.NewRecorder()
		unauthedRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
