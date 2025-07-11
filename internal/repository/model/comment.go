package model

import (
	"time"

	"github.com/google/uuid"
)

// Comment представляет модель комментария в репозиторном слое
type Comment struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	PostID    uuid.UUID  `json:"post_id" db:"post_id"`
	ParentID  *uuid.UUID `json:"parent_id" db:"parent_id"`
	Content   string     `json:"content" db:"content"`
	AuthorID  uuid.UUID  `json:"author_id" db:"author_id"`
	Depth     int        `json:"depth" db:"depth"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// CommentFilter представляет фильтры для поиска комментариев в репозитории
type CommentFilter struct {
	PostID   *uuid.UUID `json:"post_id,omitempty"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	AuthorID *uuid.UUID `json:"author_id,omitempty"`
	MaxDepth *int       `json:"max_depth,omitempty"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
	OrderBy  string     `json:"order_by"`  // "created_at", "depth"
	OrderDir string     `json:"order_dir"` // "asc", "desc"
}

// CommentWithChildren расширяет Comment информацией о дочерних комментариях
type CommentWithChildren struct {
	Comment
	ChildrenCount int `json:"children_count" db:"children_count"`
}
