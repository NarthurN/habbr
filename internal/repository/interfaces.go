package repository

import (
	"context"
	"errors"

	repomodel "github.com/NarthurN/habbr/internal/repository/model"
	"github.com/google/uuid"
)

// Общие ошибки репозиториев
var (
	ErrNotFound = errors.New("entity not found")
)

//go:generate mockery --name PostRepository --output ./mocks --filename mock_post_repository.go
type PostRepository interface {
	// Создание поста
	Create(ctx context.Context, post *repomodel.Post) error

	// Получение поста по ID
	GetByID(ctx context.Context, id uuid.UUID) (*repomodel.Post, error)

	// Получение списка постов с фильтрацией и пагинацией
	List(ctx context.Context, filter repomodel.PostFilter) ([]*repomodel.Post, error)

	// Подсчет общего количества постов с фильтрацией
	Count(ctx context.Context, filter repomodel.PostFilter) (int, error)

	// Обновление поста
	Update(ctx context.Context, post *repomodel.Post) error

	// Удаление поста
	Delete(ctx context.Context, id uuid.UUID) error

	// Проверка существования поста
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Получение постов с количеством комментариев
	ListWithCommentCounts(ctx context.Context, filter repomodel.PostFilter) ([]*repomodel.PostWithCommentCount, error)
}

//go:generate mockery --name CommentRepository --output ./mocks --filename mock_comment_repository.go
type CommentRepository interface {
	// Создание комментария
	Create(ctx context.Context, comment *repomodel.Comment) error

	// Получение комментария по ID
	GetByID(ctx context.Context, id uuid.UUID) (*repomodel.Comment, error)

	// Получение списка комментариев с фильтрацией и пагинацией
	List(ctx context.Context, filter repomodel.CommentFilter) ([]*repomodel.Comment, error)

	// Подсчет общего количества комментариев с фильтрацией
	Count(ctx context.Context, filter repomodel.CommentFilter) (int, error)

	// Обновление комментария
	Update(ctx context.Context, comment *repomodel.Comment) error

	// Удаление комментария
	Delete(ctx context.Context, id uuid.UUID) error

	// Проверка существования комментария
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Получение комментариев к посту (для построения дерева)
	GetByPostID(ctx context.Context, postID uuid.UUID) ([]*repomodel.Comment, error)

	// Получение дочерних комментариев
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*repomodel.Comment, error)

	// Получение максимальной глубины комментария
	GetMaxDepthForPost(ctx context.Context, postID uuid.UUID) (int, error)

	// Удаление всех комментариев к посту
	DeleteByPostID(ctx context.Context, postID uuid.UUID) error

	// Получение количества комментариев к посту
	CountByPostID(ctx context.Context, postID uuid.UUID) (int, error)
}

// Repositories объединяет все репозитории
type Repositories struct {
	Post    PostRepository
	Comment CommentRepository
}

// RepositoryManager управляет подключениями к репозиториям
type RepositoryManager interface {
	// Получение всех репозиториев
	GetRepositories() *Repositories

	// Закрытие соединений
	Close(ctx context.Context) error

	// Проверка здоровья соединения
	HealthCheck(ctx context.Context) error

	// Выполнение миграций (только для PostgreSQL)
	Migrate(ctx context.Context) error
}
