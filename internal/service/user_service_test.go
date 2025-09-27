package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/alfynf/simple-blog-api/internal/config"
	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/alfynf/simple-blog-api/internal/repository"
	"github.com/alfynf/simple-blog-api/internal/repository/mocks"
	"github.com/alfynf/simple-blog-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	// Create a dummy config for testing purposes
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	userService := service.NewUserService(mockUserRepo, cfg)

	ctx := context.Background()
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	t.Run("Success", func(t *testing.T) {
		// We expect the Create method to be called once with any user object.
		// The user object will have a new UUID and a hashed password.
		mockUserRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, user *model.User) error {
			// Assertions on the user object passed to the repository
			assert.Equal(t, username, user.Username)
			assert.Equal(t, email, user.Email)
			assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)), "password should be hashed correctly")
			assert.NotEqual(t, uuid.Nil, user.ID)
			return nil
		})

		user, err := userService.Register(ctx, username, email, password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, username, user.Username)
	})

	t.Run("Email or Username Already Exists", func(t *testing.T) {
		// Mock the repository to return a duplicate entry error
		mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(repository.ErrDuplicateEntry)

		_, err := userService.Register(ctx, username, email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrEmailExists)
	})

	t.Run("Failed to Create User", func(t *testing.T) {
		dbErr := errors.New("some database error")
		// Mock the repository to return a generic database error
		mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(dbErr)

		_, err := userService.Register(ctx, username, email, password)

		assert.Error(t, err)
		// Check that the original error is wrapped
		assert.Contains(t, err.Error(), "failed to create user")
		assert.ErrorIs(t, err, dbErr)
	})

}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	cfg := &config.Config{
		JWTSecret: "a-very-secret-key",
	}
	userService := service.NewUserService(mockUserRepo, cfg)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPassword),
	}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().FindByEmail(ctx, email).Return(user, nil)

		token, err := userService.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockUserRepo.EXPECT().FindByEmail(ctx, email).Return(nil, sql.ErrNoRows)

		_, err := userService.Login(ctx, email, password)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, service.ErrInvalidCredentials))
	})

	t.Run("Incorrect Password", func(t *testing.T) {
		wrongPassword := "wrongpassword"
		mockUserRepo.EXPECT().FindByEmail(ctx, email).Return(user, nil)

		_, err := userService.Login(ctx, email, wrongPassword)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, service.ErrInvalidCredentials))
	})

	t.Run("Repository Error on FindByEmail", func(t *testing.T) {
		repoErr := errors.New("some database error")
		mockUserRepo.EXPECT().FindByEmail(ctx, email).Return(nil, repoErr)

		_, err := userService.Login(ctx, email, password)

		assert.Error(t, err)
		assert.False(t, errors.Is(err, service.ErrInvalidCredentials))
		assert.Contains(t, err.Error(), "failed to find user by email")
	})
}
