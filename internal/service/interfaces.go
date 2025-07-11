package service

import (
	"context"

	"github.com/NarthurN/habbr/internal/model"
	"github.com/google/uuid"
)

//go:generate mockery --name PostService --output ./mocks --filename mock_post_service.go
type PostService interface {
	// Создание поста
	CreatePost(ctx context.Context, input model.PostInput) (*model.Post, error)

	// Получение поста по ID
	GetPost(ctx context.Context, id uuid.UUID) (*model.Post, error)

	// Получение списка постов с пагинацией
	ListPosts(ctx context.Context, filter model.PostFilter, pagination model.PaginationInput) (*model.PostConnection, error)

	// Обновление поста
	UpdatePost(ctx context.Context, id uuid.UUID, input model.PostUpdateInput, authorID uuid.UUID) (*model.Post, error)

	// Удаление поста
	DeletePost(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error

	// Переключение возможности комментирования
	ToggleComments(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, enabled bool) (*model.Post, error)
}

//go:generate mockery --name CommentService --output ./mocks --filename mock_comment_service.go
type CommentService interface {
	// Создание комментария
	CreateComment(ctx context.Context, input model.CommentInput) (*model.Comment, error)

	// Получение комментария по ID
	GetComment(ctx context.Context, id uuid.UUID) (*model.Comment, error)

	// Получение списка комментариев с пагинацией
	ListComments(ctx context.Context, filter model.CommentFilter, pagination model.PaginationInput) (*model.CommentConnection, error)

	// Получение дерева комментариев к посту
	GetCommentsTree(ctx context.Context, postID uuid.UUID) ([]*model.Comment, error)

	// Обновление комментария
	UpdateComment(ctx context.Context, id uuid.UUID, input model.CommentUpdateInput, authorID uuid.UUID) (*model.Comment, error)

	// Удаление комментария
	DeleteComment(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error
}

//go:generate mockery --name SubscriptionService --output ./mocks --filename mock_subscription_service.go
type SubscriptionService interface {
	// Подписка на комментарии к посту
	SubscribeToComments(ctx context.Context, postID uuid.UUID) (<-chan *model.CommentSubscriptionPayload, error)

	// Отписка от комментариев
	Unsubscribe(ctx context.Context, postID uuid.UUID, subscriberID string) error

	// Уведомление о новом комментарии
	NotifyCommentCreated(ctx context.Context, comment *model.Comment) error

	// Уведомление об обновлении комментария
	NotifyCommentUpdated(ctx context.Context, comment *model.Comment) error

	// Уведомление об удалении комментария
	NotifyCommentDeleted(ctx context.Context, postID uuid.UUID, commentID uuid.UUID) error
}

// Services объединяет все сервисы
type Services struct {
	Post         PostService
	Comment      CommentService
	Subscription SubscriptionService
}
