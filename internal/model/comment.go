package model

import (
	"time"

	"github.com/google/uuid"
)

// Comment merepresentasikan sebuah komentar pada sebuah post
type Comment struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
