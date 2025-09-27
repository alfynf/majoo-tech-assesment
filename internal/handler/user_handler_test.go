package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
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

func TestUserHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	userHandler := handler.NewUserHandler(mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/register", userHandler.Register)

	t.Run("Success", func(t *testing.T) {
		reqBody := handler.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}
		user := &model.User{
			ID:       uuid.New(),
			Username: reqBody.Username,
			Email:    reqBody.Email,
		}

		mockUserService.EXPECT().Register(gomock.Any(), reqBody.Username, reqBody.Email, reqBody.Password).Return(user, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var respBody model.User
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, user.Username, respBody.Username)
		assert.Equal(t, user.Email, respBody.Email)
	})

	t.Run("Binding Failure", func(t *testing.T) {
		// Missing email
		reqBody := map[string]string{
			"username": "testuser",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Email Exists", func(t *testing.T) {
		reqBody := handler.RegisterRequest{
			Username: "testuser",
			Email:    "exists@example.com",
			Password: "password123",
		}
		mockUserService.EXPECT().Register(gomock.Any(), reqBody.Username, reqBody.Email, reqBody.Password).Return(nil, service.ErrEmailExists)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), service.ErrEmailExists.Error())
	})

	t.Run("Failed to register user", func(t *testing.T) {
		reqBody := handler.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}
		mockUserService.EXPECT().Register(gomock.Any(), reqBody.Username, reqBody.Email, reqBody.Password).Return(nil, errors.New("some internal error"))

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to register user")
	})
}

func TestUserHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	userHandler := handler.NewUserHandler(mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", userHandler.Login)

	t.Run("Success", func(t *testing.T) {
		reqBody := handler.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		token := "some.jwt.token"

		mockUserService.EXPECT().Login(gomock.Any(), reqBody.Email, reqBody.Password).Return(token, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respBody handler.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, token, respBody.Token)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		reqBody := handler.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		mockUserService.EXPECT().Login(gomock.Any(), reqBody.Email, reqBody.Password).Return("", service.ErrInvalidCredentials)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), service.ErrInvalidCredentials.Error())
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		reqBody := handler.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		mockUserService.EXPECT().Login(gomock.Any(), reqBody.Email, reqBody.Password).Return("", errors.New("some database error"))

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to login")
	})
}
