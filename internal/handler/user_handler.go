package handler

import (
	"errors"
	"net/http"

	"github.com/alfynf/simple-blog-api/internal/model"
	"github.com/alfynf/simple-blog-api/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(us service.UserService) *UserHandler {
	return &UserHandler{
		userService: us,
	}
}

// RegisterRequest adalah model untuk validasi input registrasi
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest adalah model untuk validasi input login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// @Summary      Register a new user
// @Description  Creates a new user account with a username, email, and password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterRequest  true  "User Registration Details"
// @Success      201   {object}  model.User       "Successfully created user"
// @Failure      400   {object}  ErrorResponse    "Invalid input"
// @Failure      409   {object}  ErrorResponse    "Email or username already exists"
// @Failure      500   {object}  ErrorResponse    "Internal server error"
// @Router       /register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	// Validasi input JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user", "details": err.Error()})
		return
	}

	// Buat response tanpa password
	userResponse := model.User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, userResponse)
}

// @Summary      Login a user
// @Description  Logs in a user with email and password, and returns a JWT token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest  true  "User Login Credentials"
// @Success      200          {object}  LoginResponse "Successfully logged in"
// @Failure      400          {object}  ErrorResponse "Invalid input"
// @Failure      401          {object}  ErrorResponse "Invalid credentials"
// @Failure      500          {object}  ErrorResponse "Internal server error"
// @Router       /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
