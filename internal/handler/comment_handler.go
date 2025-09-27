package handler

import (
	"errors"
	"net/http"

	"github.com/alfynf/simple-blog-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CommentHandler struct {
	commentService service.CommentService
}

func NewCommentHandler(cs service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: cs,
	}
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

// @Summary      Create a new comment on a post
// @Description  Adds a new comment to a specific blog post. Requires authentication.
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id      path      string                true  "Post ID"
// @Param        comment body      CreateCommentRequest  true  "Comment content"
// @Success      201     {object}  model.Comment
// @Failure      400     {object}  ErrorResponse "Invalid input or ID format"
// @Failure      401     {object}  ErrorResponse "Unauthorized"
// @Failure      404     {object}  ErrorResponse "Post not found"
// @Failure      500     {object}  ErrorResponse "Internal server error"
// @Router       /posts/{id}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid post ID format"})
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	comment, err := h.commentService.CreateComment(c.Request.Context(), userID.(uuid.UUID), postID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// @Summary      Get all comments for a post
// @Description  Retrieves all comments associated with a specific blog post.
// @Tags         Comments
// @Produce      json
// @Param        id   path      string  true  "Post ID"
// @Success      200  {array}   model.Comment
// @Failure      400  {object}  ErrorResponse "Invalid post ID format"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /posts/{id}/comments [get]
func (h *CommentHandler) GetCommentsByPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid post ID format"})
		return
	}

	comments, err := h.commentService.GetCommentsByPostID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// @Summary      Update a comment
// @Description  Updates an existing comment. Requires authentication and ownership.
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        commentID  path      string                true  "Comment ID"
// @Param        comment    body      UpdateCommentRequest  true  "Updated comment content"
// @Success      200        {object}  model.Comment
// @Failure      400        {object}  ErrorResponse "Invalid input or ID format"
// @Failure      401        {object}  ErrorResponse "Unauthorized"
// @Failure      403        {object}  ErrorResponse "Forbidden"
// @Failure      404        {object}  ErrorResponse "Comment not found"
// @Failure      500        {object}  ErrorResponse "Internal server error"
// @Router       /comments/{commentID} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("commentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid comment ID format"})
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	comment, err := h.commentService.UpdateComment(c.Request.Context(), commentID, userID.(uuid.UUID), req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update comment"})
		}
		return
	}

	c.JSON(http.StatusOK, comment)
}

// @Summary      Delete a comment
// @Description  Deletes a comment. Requires authentication and ownership.
// @Tags         Comments
// @Produce      json
// @Security     ApiKeyAuth
// @Param        commentID  path      string  true  "Comment ID"
// @Success      204        "No Content"
// @Failure      400        {object}  ErrorResponse "Invalid comment ID format"
// @Failure      401        {object}  ErrorResponse "Unauthorized"
// @Failure      404        {object}  ErrorResponse "Comment not found or not owned by user"
// @Failure      500        {object}  ErrorResponse "Internal server error"
// @Router       /comments/{commentID} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("commentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid comment ID format"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	err = h.commentService.DeleteComment(c.Request.Context(), commentID, userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) || errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Comment not found or not owned by user"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete comment"})
		return
	}

	c.Status(http.StatusNoContent)
}
