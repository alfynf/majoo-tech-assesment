package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/alfynf/simple-blog-api/internal/config"
	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/alfynf/simple-blog-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email or username already exists")
)

// UserService mendefinisikan logika bisnis untuk User
type UserService interface {
	Register(ctx context.Context, username, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewUserService adalah constructor untuk userService
func NewUserService(repo repository.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: repo,
		cfg:      cfg,
	}
}

// Register menghandle validasi dan pembuatan user baru
func (s *userService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	// 1. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 2. Buat model user
	user := &model.User{
		ID:       uuid.New(),
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	// 3. Simpan ke database melalui repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEntry) {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login memverifikasi kredensial user dan mengembalikan JWT
func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Cari user berdasarkan email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to find user by email: %w", err)
	}

	// 2. Bandingkan password yang di-hash dengan password yang diberikan
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Jika error adalah password tidak cocok, kembalikan error kredensial
		return "", ErrInvalidCredentials
	}

	// 3. Buat JWT Claims
	claims := jwt.MapClaims{
		"sub": user.ID,                               // Subject (user ID)
		"exp": time.Now().Add(time.Hour * 72).Unix(), // Expiration time (e.g., 72 hours)
		"iat": time.Now().Unix(),                     // Issued at
	}

	// 4. Buat token dengan claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 5. Tanda tangani token dengan secret key
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
