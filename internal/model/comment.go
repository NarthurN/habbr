package model

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxCommentLength = 2000

// Comment представляет доменную модель комментария
type Comment struct {
	ID        uuid.UUID  `json:"id"`
	PostID    uuid.UUID  `json:"post_id"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Content   string     `json:"content"`
	AuthorID  uuid.UUID  `json:"author_id"`
	Depth     int        `json:"depth"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Children  []*Comment `json:"children,omitempty"`
}

// CommentInput представляет входные данные для создания комментария
type CommentInput struct {
	PostID   uuid.UUID  `json:"post_id"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Content  string     `json:"content"`
	AuthorID uuid.UUID  `json:"author_id"`
}

// CommentUpdateInput представляет входные данные для обновления комментария
type CommentUpdateInput struct {
	Content *string `json:"content,omitempty"`
}

// CommentFilter представляет фильтры для поиска комментариев
type CommentFilter struct {
	PostID   *uuid.UUID `json:"post_id,omitempty"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	AuthorID *uuid.UUID `json:"author_id,omitempty"`
	MaxDepth *int       `json:"max_depth,omitempty"`
}

// CommentConnection представляет связь комментариев с пагинацией
type CommentConnection struct {
	Edges    []*CommentEdge `json:"edges"`
	PageInfo *PageInfo      `json:"page_info"`
}

// CommentEdge представляет ребро комментария в пагинации
type CommentEdge struct {
	Node   *Comment `json:"node"`
	Cursor string   `json:"cursor"`
}

// CommentSubscriptionPayload представляет данные для подписки на комментарии
type CommentSubscriptionPayload struct {
	PostID     uuid.UUID `json:"post_id"`
	Comment    *Comment  `json:"comment"`
	ActionType string    `json:"action_type"` // "CREATED", "UPDATED", "DELETED"
}

// Validate проверяет валидность данных комментария
func (c *CommentInput) Validate() error {
	if strings.TrimSpace(c.Content) == "" {
		return errors.New("content cannot be empty")
	}

	if len(c.Content) > MaxCommentLength {
		return errors.New("content cannot exceed 2000 characters")
	}

	if c.PostID == uuid.Nil {
		return errors.New("post_id is required")
	}

	if c.AuthorID == uuid.Nil {
		return errors.New("author_id is required")
	}

	return nil
}

// Validate проверяет валидность данных для обновления комментария
func (c *CommentUpdateInput) Validate() error {
	if c.Content != nil {
		if strings.TrimSpace(*c.Content) == "" {
			return errors.New("content cannot be empty")
		}

		if len(*c.Content) > MaxCommentLength {
			return errors.New("content cannot exceed 2000 characters")
		}
	}

	return nil
}

// NewComment создает новый комментарий из входных данных
func NewComment(input CommentInput, depth int) *Comment {
	now := time.Now()
	return &Comment{
		ID:        uuid.New(),
		PostID:    input.PostID,
		ParentID:  input.ParentID,
		Content:   strings.TrimSpace(input.Content),
		AuthorID:  input.AuthorID,
		Depth:     depth,
		CreatedAt: now,
		UpdatedAt: now,
		Children:  make([]*Comment, 0),
	}
}

// Update обновляет комментарий новыми данными
func (c *Comment) Update(input CommentUpdateInput) {
	if input.Content != nil {
		c.Content = strings.TrimSpace(*input.Content)
	}

	c.UpdatedAt = time.Now()
}

// IsRootComment проверяет, является ли комментарий корневым
func (c *Comment) IsRootComment() bool {
	return c.ParentID == nil
}

// CanBeRepliedTo проверяет, можно ли ответить на комментарий
func (c *Comment) CanBeRepliedTo() bool {
	// Можно добавить ограничения на глубину вложенности если потребуется
	return true
}

// AddChild добавляет дочерний комментарий
func (c *Comment) AddChild(child *Comment) {
	if c.Children == nil {
		c.Children = make([]*Comment, 0)
	}
	c.Children = append(c.Children, child)
}

// GetDepth возвращает глубину комментария в дереве
func (c *Comment) GetDepth() int {
	return c.Depth
}

// BuildCommentsTree строит дерево комментариев из плоского списка
func BuildCommentsTree(comments []*Comment) []*Comment {
	if len(comments) == 0 {
		return make([]*Comment, 0)
	}

	// Создаем карту для быстрого поиска комментариев по ID
	commentMap := make(map[uuid.UUID]*Comment)
	for _, comment := range comments {
		comment.Children = make([]*Comment, 0) // Инициализируем дочерние комментарии
		commentMap[comment.ID] = comment
	}

	// Строим дерево
	rootComments := make([]*Comment, 0)
	for _, comment := range comments {
		if comment.IsRootComment() {
			rootComments = append(rootComments, comment)
		} else if parent, exists := commentMap[*comment.ParentID]; exists {
			parent.AddChild(comment)
		}
	}

	return rootComments
}

// FlattenCommentsTree преобразует дерево комментариев в плоский список
func FlattenCommentsTree(tree []*Comment) []*Comment {
	result := make([]*Comment, 0)

	var flatten func([]*Comment)
	flatten = func(comments []*Comment) {
		for _, comment := range comments {
			result = append(result, comment)
			if len(comment.Children) > 0 {
				flatten(comment.Children)
			}
		}
	}

	flatten(tree)
	return result
}
