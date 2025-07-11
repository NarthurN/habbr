package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	repomodel "github.com/NarthurN/habbr/internal/repository/model"
	"github.com/google/uuid"
)

// CommentRepository представляет in-memory реализацию репозитория комментариев
type CommentRepository struct {
	mu       sync.RWMutex
	comments map[uuid.UUID]*repomodel.Comment
}

// NewCommentRepository создает новый in-memory репозиторий комментариев
func NewCommentRepository() *CommentRepository {
	return &CommentRepository{
		comments: make(map[uuid.UUID]*repomodel.Comment),
	}
}

// Create создает новый комментарий
func (r *CommentRepository) Create(ctx context.Context, comment *repomodel.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comment == nil {
		return fmt.Errorf("comment cannot be nil")
	}

	// Проверяем, что комментарий с таким ID не существует
	if _, exists := r.comments[comment.ID]; exists {
		return fmt.Errorf("comment with ID %s already exists", comment.ID)
	}

	// Создаем копию комментария
	commentCopy := *comment
	r.comments[comment.ID] = &commentCopy

	return nil
}

// GetByID возвращает комментарий по ID
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*repomodel.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	comment, exists := r.comments[id]
	if !exists {
		return nil, nil // не найден
	}

	// Возвращаем копию
	commentCopy := *comment
	return &commentCopy, nil
}

// List возвращает список комментариев с фильтрацией и пагинацией
func (r *CommentRepository) List(ctx context.Context, filter repomodel.CommentFilter) ([]*repomodel.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Собираем все комментарии
	allComments := make([]*repomodel.Comment, 0, len(r.comments))
	for _, comment := range r.comments {
		// Применяем фильтры
		if filter.PostID != nil && comment.PostID != *filter.PostID {
			continue
		}

		if filter.ParentID != nil &&
			((comment.ParentID == nil && *filter.ParentID != uuid.Nil) ||
				(comment.ParentID != nil && *comment.ParentID != *filter.ParentID)) {
			continue
		}

		if filter.AuthorID != nil && comment.AuthorID != *filter.AuthorID {
			continue
		}

		if filter.MaxDepth != nil && comment.Depth > *filter.MaxDepth {
			continue
		}

		// Создаем копию
		commentCopy := *comment
		allComments = append(allComments, &commentCopy)
	}

	// Сортируем
	r.sortComments(allComments, filter.OrderBy, filter.OrderDir)

	// Применяем пагинацию
	start := filter.Offset
	end := start + filter.Limit

	if start >= len(allComments) {
		return []*repomodel.Comment{}, nil
	}

	if end > len(allComments) {
		end = len(allComments)
	}

	return allComments[start:end], nil
}

// Count возвращает общее количество комментариев с фильтрацией
func (r *CommentRepository) Count(ctx context.Context, filter repomodel.CommentFilter) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, comment := range r.comments {
		// Применяем фильтры
		if filter.PostID != nil && comment.PostID != *filter.PostID {
			continue
		}

		if filter.ParentID != nil &&
			((comment.ParentID == nil && *filter.ParentID != uuid.Nil) ||
				(comment.ParentID != nil && *comment.ParentID != *filter.ParentID)) {
			continue
		}

		if filter.AuthorID != nil && comment.AuthorID != *filter.AuthorID {
			continue
		}

		if filter.MaxDepth != nil && comment.Depth > *filter.MaxDepth {
			continue
		}

		count++
	}

	return count, nil
}

// Update обновляет комментарий
func (r *CommentRepository) Update(ctx context.Context, comment *repomodel.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comment == nil {
		return fmt.Errorf("comment cannot be nil")
	}

	if _, exists := r.comments[comment.ID]; !exists {
		return fmt.Errorf("comment with ID %s not found", comment.ID)
	}

	// Обновляем время изменения
	comment.UpdatedAt = time.Now()

	// Создаем копию и сохраняем
	commentCopy := *comment
	r.comments[comment.ID] = &commentCopy

	return nil
}

// Delete удаляет комментарий
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.comments[id]; !exists {
		return fmt.Errorf("comment with ID %s not found", id)
	}

	delete(r.comments, id)
	return nil
}

// Exists проверяет существование комментария
func (r *CommentRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.comments[id]
	return exists, nil
}

// GetByPostID возвращает все комментарии к посту
func (r *CommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID) ([]*repomodel.Comment, error) {
	filter := repomodel.CommentFilter{
		PostID:   &postID,
		Limit:    1000, // большой лимит для получения всех комментариев
		Offset:   0,
		OrderBy:  "created_at",
		OrderDir: "asc",
	}

	return r.List(ctx, filter)
}

// GetChildren возвращает дочерние комментарии
func (r *CommentRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*repomodel.Comment, error) {
	filter := repomodel.CommentFilter{
		ParentID: &parentID,
		Limit:    1000,
		Offset:   0,
		OrderBy:  "created_at",
		OrderDir: "asc",
	}

	return r.List(ctx, filter)
}

// GetMaxDepthForPost возвращает максимальную глубину комментариев к посту
func (r *CommentRepository) GetMaxDepthForPost(ctx context.Context, postID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	maxDepth := 0
	for _, comment := range r.comments {
		if comment.PostID == postID && comment.Depth > maxDepth {
			maxDepth = comment.Depth
		}
	}

	return maxDepth, nil
}

// DeleteByPostID удаляет все комментарии к посту
func (r *CommentRepository) DeleteByPostID(ctx context.Context, postID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Собираем ID комментариев для удаления
	var idsToDelete []uuid.UUID
	for id, comment := range r.comments {
		if comment.PostID == postID {
			idsToDelete = append(idsToDelete, id)
		}
	}

	// Удаляем комментарии
	for _, id := range idsToDelete {
		delete(r.comments, id)
	}

	return nil
}

// CountByPostID возвращает количество комментариев к посту
func (r *CommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int, error) {
	filter := repomodel.CommentFilter{
		PostID: &postID,
	}

	return r.Count(ctx, filter)
}

// sortComments сортирует комментарии по указанному полю и направлению
func (r *CommentRepository) sortComments(comments []*repomodel.Comment, orderBy, orderDir string) {
	if orderBy == "" {
		orderBy = "created_at"
	}

	if orderDir == "" {
		orderDir = "asc"
	}

	sort.Slice(comments, func(i, j int) bool {
		var result bool

		switch orderBy {
		case "depth":
			if comments[i].Depth == comments[j].Depth {
				// Если глубина одинаковая, сортируем по времени создания
				result = comments[i].CreatedAt.Before(comments[j].CreatedAt)
			} else {
				result = comments[i].Depth < comments[j].Depth
			}
		default: // "created_at"
			result = comments[i].CreatedAt.Before(comments[j].CreatedAt)
		}

		if orderDir == "desc" {
			result = !result
		}

		return result
	})
}
