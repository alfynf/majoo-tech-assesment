package model

import (
	"time"

	"github.com/google/uuid"
)

// Post merepresentasikan sebuah postingan blog
type Post struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"` // ID pengguna yang membuat postingan
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
