package handler

import (
	"errors"
	"net/http"

	"github.com/alfynf/simple-blog-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostHandler struct {
	postService service.PostService
}

func NewPostHandler(ps service.PostService) *PostHandler {
	return &PostHandler{
		postService: ps,
	}
}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=5"`
	Content string `json:"content" binding:"required,min=10"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"required,min=5"`
	Content string `json:"content" binding:"required,min=10"`
}

// @Summary      Create a new post
// @Description  Creates a new blog post. Requires authentication.
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        post  body      CreatePostRequest  true  "Post details"
// @Success      201   {object}  model.Post
// @Failure      400   {object}  ErrorResponse "Invalid input"
// @Failure      401   {object}  ErrorResponse "Unauthorized"
// @Failure      500   {object}  ErrorResponse "Internal server error"
// @Router       /posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	post, err := h.postService.CreatePost(c.Request.Context(), userID.(uuid.UUID), req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// @Summary      Get a post by ID
// @Description  Retrieves the details of a single blog post by its ID.
// @Tags         Posts
// @Produce      json
// @Param        id   path      string  true  "Post ID"
// @Success      200  {object}  model.Post
// @Failure      400  {object}  ErrorResponse "Invalid post ID format"
// @Failure      404  {object}  ErrorResponse "Post not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid post ID format"})
		return
	}

	post, err := h.postService.GetPostByID(c.Request.Context(), postID)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve post"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// @Summary      Update a post
// @Description  Updates an existing blog post. Requires authentication and ownership.
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string             true  "Post ID"
// @Param        post body      UpdatePostRequest  true  "Updated post details"
// @Success      200  {object}  model.Post
// @Failure      400  {object}  ErrorResponse "Invalid input or ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized"
// @Failure      403  {object}  ErrorResponse "Forbidden"
// @Failure      404  {object}  ErrorResponse "Post not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid post ID format"})
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), postID, userID.(uuid.UUID), req.Title, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update post"})
		}
		return
	}

	c.JSON(http.StatusOK, post)
}

// @Summary      Delete a post
// @Description  Deletes a blog post. Requires authentication and ownership.
// @Tags         Posts
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id  path      string  true  "Post ID"
// @Success      204 "No Content"
// @Failure      400 {object}  ErrorResponse "Invalid post ID format"
// @Failure      401 {object}  ErrorResponse "Unauthorized"
// @Failure      403 {object}  ErrorResponse "Forbidden"
// @Failure      404 {object}  ErrorResponse "Post not found"
// @Failure      500 {object}  ErrorResponse "Internal server error"
// @Router       /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid post ID format"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	err = h.postService.DeletePost(c.Request.Context(), postID, userID.(uuid.UUID))
	if err != nil {
		// To prevent leaking information, we can return 404 for both not found and forbidden
		if errors.Is(err, service.ErrPostNotFound) || errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Post not found or not owned by user"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete post"})
		return
	}

	c.Status(http.StatusNoContent)
}
