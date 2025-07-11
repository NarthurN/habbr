package model

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Post представляет доменную модель поста
type Post struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	AuthorID        uuid.UUID `json:"author_id"`
	CommentsEnabled bool      `json:"comments_enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PostInput представляет входные данные для создания поста
type PostInput struct {
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	AuthorID        uuid.UUID `json:"author_id"`
	CommentsEnabled bool      `json:"comments_enabled"`
}

// PostUpdateInput представляет входные данные для обновления поста
type PostUpdateInput struct {
	Title           *string `json:"title,omitempty"`
	Content         *string `json:"content,omitempty"`
	CommentsEnabled *bool   `json:"comments_enabled,omitempty"`
}

// PostFilter представляет фильтры для поиска постов
type PostFilter struct {
	AuthorID     *uuid.UUID `json:"author_id,omitempty"`
	WithComments *bool      `json:"with_comments,omitempty"`
}

// PaginationInput представляет параметры пагинации
type PaginationInput struct {
	First  *int    `json:"first,omitempty"`
	After  *string `json:"after,omitempty"`
	Last   *int    `json:"last,omitempty"`
	Before *string `json:"before,omitempty"`
}

// PostConnection представляет связь постов с пагинацией
type PostConnection struct {
	Edges    []*PostEdge `json:"edges"`
	PageInfo *PageInfo   `json:"page_info"`
}

// PostEdge представляет ребро поста в пагинации
type PostEdge struct {
	Node   *Post  `json:"node"`
	Cursor string `json:"cursor"`
}

// PageInfo содержит информацию о пагинации
type PageInfo struct {
	HasNextPage     bool    `json:"has_next_page"`
	HasPreviousPage bool    `json:"has_previous_page"`
	StartCursor     *string `json:"start_cursor"`
	EndCursor       *string `json:"end_cursor"`
}

// Validate проверяет валидность данных поста
func (p *PostInput) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("title cannot be empty")
	}

	if len(p.Title) > 200 {
		return errors.New("title cannot exceed 200 characters")
	}

	if strings.TrimSpace(p.Content) == "" {
		return errors.New("content cannot be empty")
	}

	if len(p.Content) > 50000 {
		return errors.New("content cannot exceed 50000 characters")
	}

	if p.AuthorID == uuid.Nil {
		return errors.New("author_id is required")
	}

	return nil
}

// Validate проверяет валидность данных для обновления поста
func (p *PostUpdateInput) Validate() error {
	if p.Title != nil {
		if strings.TrimSpace(*p.Title) == "" {
			return errors.New("title cannot be empty")
		}

		if len(*p.Title) > 200 {
			return errors.New("title cannot exceed 200 characters")
		}
	}

	if p.Content != nil {
		if strings.TrimSpace(*p.Content) == "" {
			return errors.New("content cannot be empty")
		}

		if len(*p.Content) > 50000 {
			return errors.New("content cannot exceed 50000 characters")
		}
	}

	return nil
}

// NewPost создает новый пост из входных данных
func NewPost(input PostInput) *Post {
	now := time.Now()
	return &Post{
		ID:              uuid.New(),
		Title:           strings.TrimSpace(input.Title),
		Content:         strings.TrimSpace(input.Content),
		AuthorID:        input.AuthorID,
		CommentsEnabled: input.CommentsEnabled,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Update обновляет пост новыми данными
func (p *Post) Update(input PostUpdateInput) {
	if input.Title != nil {
		p.Title = strings.TrimSpace(*input.Title)
	}

	if input.Content != nil {
		p.Content = strings.TrimSpace(*input.Content)
	}

	if input.CommentsEnabled != nil {
		p.CommentsEnabled = *input.CommentsEnabled
	}

	p.UpdatedAt = time.Now()
}

// CanAddComments проверяет, можно ли добавлять комментарии к посту
func (p *Post) CanAddComments() bool {
	return p.CommentsEnabled
}
