package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	repomodel "github.com/NarthurN/habbr/internal/repository/model"
	"github.com/google/uuid"
)

// PostRepository представляет in-memory реализацию репозитория постов
type PostRepository struct {
	mu    sync.RWMutex
	posts map[uuid.UUID]*repomodel.Post
}

// NewPostRepository создает новый in-memory репозиторий постов
func NewPostRepository() *PostRepository {
	return &PostRepository{
		posts: make(map[uuid.UUID]*repomodel.Post),
	}
}

// Create создает новый пост
func (r *PostRepository) Create(ctx context.Context, post *repomodel.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if post == nil {
		return fmt.Errorf("post cannot be nil")
	}

	// Проверяем, что пост с таким ID не существует
	if _, exists := r.posts[post.ID]; exists {
		return fmt.Errorf("post with ID %s already exists", post.ID)
	}

	// Создаем копию поста
	postCopy := *post
	r.posts[post.ID] = &postCopy

	return nil
}

// GetByID возвращает пост по ID
func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*repomodel.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.posts[id]
	if !exists {
		return nil, nil // не найден
	}

	// Возвращаем копию
	postCopy := *post
	return &postCopy, nil
}

// List возвращает список постов с фильтрацией и пагинацией
func (r *PostRepository) List(ctx context.Context, filter repomodel.PostFilter) ([]*repomodel.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Собираем все посты
	allPosts := make([]*repomodel.Post, 0, len(r.posts))
	for _, post := range r.posts {
		// Применяем фильтры
		if filter.AuthorID != nil && post.AuthorID != *filter.AuthorID {
			continue
		}

		if filter.WithComments != nil {
			// Для in-memory реализации просто игнорируем этот фильтр
			// В реальной реализации нужно было бы проверить наличие комментариев
		}

		// Создаем копию
		postCopy := *post
		allPosts = append(allPosts, &postCopy)
	}

	// Сортируем
	r.sortPosts(allPosts, filter.OrderBy, filter.OrderDir)

	// Применяем пагинацию
	start := filter.Offset
	end := start + filter.Limit

	if start >= len(allPosts) {
		return []*repomodel.Post{}, nil
	}

	if end > len(allPosts) {
		end = len(allPosts)
	}

	return allPosts[start:end], nil
}

// Count возвращает общее количество постов с фильтрацией
func (r *PostRepository) Count(ctx context.Context, filter repomodel.PostFilter) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, post := range r.posts {
		// Применяем фильтры
		if filter.AuthorID != nil && post.AuthorID != *filter.AuthorID {
			continue
		}

		if filter.WithComments != nil {
			// Для in-memory реализации просто игнорируем этот фильтр
		}

		count++
	}

	return count, nil
}

// Update обновляет пост
func (r *PostRepository) Update(ctx context.Context, post *repomodel.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if post == nil {
		return fmt.Errorf("post cannot be nil")
	}

	if _, exists := r.posts[post.ID]; !exists {
		return fmt.Errorf("post with ID %s not found", post.ID)
	}

	// Обновляем время изменения
	post.UpdatedAt = time.Now()

	// Создаем копию и сохраняем
	postCopy := *post
	r.posts[post.ID] = &postCopy

	return nil
}

// Delete удаляет пост
func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[id]; !exists {
		return fmt.Errorf("post with ID %s not found", id)
	}

	delete(r.posts, id)
	return nil
}

// Exists проверяет существование поста
func (r *PostRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.posts[id]
	return exists, nil
}

// ListWithCommentCounts возвращает посты с количеством комментариев
func (r *PostRepository) ListWithCommentCounts(ctx context.Context, filter repomodel.PostFilter) ([]*repomodel.PostWithCommentCount, error) {
	// Получаем обычный список постов
	posts, err := r.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Конвертируем в PostWithCommentCount
	result := make([]*repomodel.PostWithCommentCount, len(posts))
	for i, post := range posts {
		result[i] = &repomodel.PostWithCommentCount{
			Post:         *post,
			CommentCount: 0, // Для in-memory реализации не считаем комментарии
		}
	}

	return result, nil
}

// sortPosts сортирует посты по указанному полю и направлению
func (r *PostRepository) sortPosts(posts []*repomodel.Post, orderBy, orderDir string) {
	if orderBy == "" {
		orderBy = "created_at"
	}

	if orderDir == "" {
		orderDir = "desc"
	}

	sort.Slice(posts, func(i, j int) bool {
		var result bool

		switch orderBy {
		case "title":
			result = strings.Compare(posts[i].Title, posts[j].Title) < 0
		case "updated_at":
			result = posts[i].UpdatedAt.Before(posts[j].UpdatedAt)
		default: // "created_at"
			result = posts[i].CreatedAt.Before(posts[j].CreatedAt)
		}

		if orderDir == "desc" {
			result = !result
		}

		return result
	})
}
