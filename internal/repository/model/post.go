package model

import (
	"time"

	"github.com/google/uuid"
)

// Post представляет модель поста в репозиторном слое
type Post struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Content         string    `json:"content" db:"content"`
	AuthorID        uuid.UUID `json:"author_id" db:"author_id"`
	CommentsEnabled bool      `json:"comments_enabled" db:"comments_enabled"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// PostFilter представляет фильтры для поиска постов в репозитории
type PostFilter struct {
	AuthorID     *uuid.UUID `json:"author_id,omitempty"`
	WithComments *bool      `json:"with_comments,omitempty"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
	OrderBy      string     `json:"order_by"`  // "created_at", "updated_at", "title"
	OrderDir     string     `json:"order_dir"` // "asc", "desc"
}

// PostWithCommentCount расширяет Post информацией о количестве комментариев
type PostWithCommentCount struct {
	Post
	CommentCount int `json:"comment_count" db:"comment_count"`
}
