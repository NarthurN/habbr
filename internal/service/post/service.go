package post

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/NarthurN/habbr/internal/model"
	"github.com/NarthurN/habbr/internal/repository"
	"github.com/NarthurN/habbr/internal/repository/converter"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service реализует бизнес-логику для работы с постами
type Service struct {
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
	logger      *zap.Logger
}

// NewService создает новый сервис постов
func NewService(repos *repository.Repositories, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Service{
		postRepo:    repos.Post,
		commentRepo: repos.Comment,
		logger:      logger,
	}
}

// CreatePost создает новый пост
func (s *Service) CreatePost(ctx context.Context, input model.PostInput) (*model.Post, error) {
	s.logger.Debug("Creating new post",
		zap.String("title", input.Title),
		zap.String("author_id", input.AuthorID.String()),
		zap.Bool("comments_enabled", input.CommentsEnabled),
	)

	// Валидация входных данных
	if err := input.Validate(); err != nil {
		s.logger.Warn("Post validation failed",
			zap.Error(err),
			zap.String("author_id", input.AuthorID.String()),
		)
		return nil, model.NewValidationError("input", err.Error())
	}

	// Создание доменной модели
	post := model.NewPost(input)

	// Конвертация в модель репозитория и сохранение
	repoPost := converter.PostToRepo(post)
	if err := s.postRepo.Create(ctx, repoPost); err != nil {
		s.logger.Error("Failed to create post in repository",
			zap.Error(err),
			zap.String("post_id", post.ID.String()),
			zap.String("author_id", post.AuthorID.String()),
		)

		// Проверяем тип ошибки репозитория
		if err == repository.ErrNotFound {
			return nil, model.NewNotFoundError("post", post.ID)
		}

		return nil, model.NewInternalError(fmt.Sprintf("failed to create post: %v", err))
	}

	s.logger.Info("Post created successfully",
		zap.String("post_id", post.ID.String()),
		zap.String("title", post.Title),
		zap.String("author_id", post.AuthorID.String()),
	)

	return post, nil
}

// GetPost возвращает пост по ID
func (s *Service) GetPost(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	if id == uuid.Nil {
		s.logger.Warn("Attempt to get post with nil ID")
		return nil, model.NewValidationError("id", "post ID is required")
	}

	s.logger.Debug("Getting post by ID", zap.String("post_id", id.String()))

	repoPost, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			s.logger.Debug("Post not found", zap.String("post_id", id.String()))
			return nil, model.NewNotFoundError("post", id)
		}

		s.logger.Error("Failed to get post from repository",
			zap.Error(err),
			zap.String("post_id", id.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to get post: %v", err))
	}

	post := converter.PostFromRepo(repoPost)
	s.logger.Debug("Post retrieved successfully",
		zap.String("post_id", post.ID.String()),
		zap.String("title", post.Title),
	)

	return post, nil
}

// ListPosts возвращает список постов с пагинацией
func (s *Service) ListPosts(ctx context.Context, filter model.PostFilter, pagination model.PaginationInput) (*model.PostConnection, error) {
	s.logger.Debug("Listing posts",
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
	)

	// Валидация пагинации
	if err := s.validatePagination(pagination); err != nil {
		s.logger.Warn("Invalid pagination parameters", zap.Error(err))
		return nil, err
	}

	// Конвертация фильтра
	repoFilter := converter.PostFilterToRepo(filter, pagination)

	// Получение постов
	repoPosts, err := s.postRepo.List(ctx, repoFilter)
	if err != nil {
		s.logger.Error("Failed to list posts from repository",
			zap.Error(err),
			zap.Any("filter", repoFilter),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to list posts: %v", err))
	}

	// Получение общего количества для более точной пагинации
	totalCount, err := s.postRepo.Count(ctx, repoFilter)
	if err != nil {
		s.logger.Error("Failed to count posts from repository",
			zap.Error(err),
			zap.Any("filter", repoFilter),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to count posts: %v", err))
	}

	// Конвертация в доменные модели
	posts := converter.PostsFromRepo(repoPosts)

	// Создание connection с пагинацией
	connection := s.buildPostConnection(posts, pagination, totalCount)

	s.logger.Debug("Posts listed successfully",
		zap.Int("count", len(posts)),
		zap.Int("total_count", totalCount),
		zap.Bool("has_next_page", connection.PageInfo.HasNextPage),
	)

	return connection, nil
}

// UpdatePost обновляет пост
func (s *Service) UpdatePost(ctx context.Context, id uuid.UUID, input model.PostUpdateInput, authorID uuid.UUID) (*model.Post, error) {
	s.logger.Debug("Updating post",
		zap.String("post_id", id.String()),
		zap.String("author_id", authorID.String()),
	)

	// Валидация входных данных
	if err := input.Validate(); err != nil {
		s.logger.Warn("Post update validation failed",
			zap.Error(err),
			zap.String("post_id", id.String()),
		)
		return nil, model.NewValidationError("input", err.Error())
	}

	if id == uuid.Nil {
		s.logger.Warn("Attempt to update post with nil ID")
		return nil, model.NewValidationError("id", "post ID is required")
	}

	if authorID == uuid.Nil {
		s.logger.Warn("Attempt to update post with nil author ID",
			zap.String("post_id", id.String()),
		)
		return nil, model.NewValidationError("author_id", "author ID is required")
	}

	// Получение существующего поста
	existingPost, err := s.GetPost(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверка прав на редактирование
	if existingPost.AuthorID != authorID {
		s.logger.Warn("Unauthorized attempt to update post",
			zap.String("post_id", id.String()),
			zap.String("post_author", existingPost.AuthorID.String()),
			zap.String("requesting_user", authorID.String()),
		)
		return nil, model.NewForbiddenError("update post")
	}

	// Сохранение исходных значений для логирования
	originalTitle := existingPost.Title
	originalCommentsEnabled := existingPost.CommentsEnabled

	// Обновление поста
	existingPost.Update(input)

	// Сохранение изменений
	repoPost := converter.PostToRepo(existingPost)
	if err := s.postRepo.Update(ctx, repoPost); err != nil {
		if err == repository.ErrNotFound {
			return nil, model.NewNotFoundError("post", id)
		}

		s.logger.Error("Failed to update post in repository",
			zap.Error(err),
			zap.String("post_id", id.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to update post: %v", err))
	}

	s.logger.Info("Post updated successfully",
		zap.String("post_id", id.String()),
		zap.String("author_id", authorID.String()),
		zap.String("old_title", originalTitle),
		zap.String("new_title", existingPost.Title),
		zap.Bool("old_comments_enabled", originalCommentsEnabled),
		zap.Bool("new_comments_enabled", existingPost.CommentsEnabled),
	)

	return existingPost, nil
}

// DeletePost удаляет пост
func (s *Service) DeletePost(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error {
	s.logger.Debug("Deleting post",
		zap.String("post_id", id.String()),
		zap.String("author_id", authorID.String()),
	)

	if id == uuid.Nil {
		s.logger.Warn("Attempt to delete post with nil ID")
		return model.NewValidationError("id", "post ID is required")
	}

	if authorID == uuid.Nil {
		s.logger.Warn("Attempt to delete post with nil author ID",
			zap.String("post_id", id.String()),
		)
		return model.NewValidationError("author_id", "author ID is required")
	}

	// Получение поста для проверки прав
	post, err := s.GetPost(ctx, id)
	if err != nil {
		return err
	}

	// Проверка прав на удаление
	if post.AuthorID != authorID {
		s.logger.Warn("Unauthorized attempt to delete post",
			zap.String("post_id", id.String()),
			zap.String("post_author", post.AuthorID.String()),
			zap.String("requesting_user", authorID.String()),
		)
		return model.NewForbiddenError("delete post")
	}

	// Подсчет комментариев для логирования
	commentCount, err := s.commentRepo.CountByPostID(ctx, id)
	if err != nil {
		s.logger.Warn("Failed to count comments before post deletion",
			zap.Error(err),
			zap.String("post_id", id.String()),
		)
		// Не прерываем выполнение, это не критично
		commentCount = 0
	}

	// Удаление всех комментариев к посту
	if err := s.commentRepo.DeleteByPostID(ctx, id); err != nil {
		s.logger.Error("Failed to delete post comments",
			zap.Error(err),
			zap.String("post_id", id.String()),
		)
		return model.NewInternalError(fmt.Sprintf("failed to delete post comments: %v", err))
	}

	// Удаление поста
	if err := s.postRepo.Delete(ctx, id); err != nil {
		if err == repository.ErrNotFound {
			return model.NewNotFoundError("post", id)
		}

		s.logger.Error("Failed to delete post from repository",
			zap.Error(err),
			zap.String("post_id", id.String()),
		)
		return model.NewInternalError(fmt.Sprintf("failed to delete post: %v", err))
	}

	s.logger.Info("Post deleted successfully",
		zap.String("post_id", id.String()),
		zap.String("title", post.Title),
		zap.String("author_id", authorID.String()),
		zap.Int("deleted_comments", commentCount),
	)

	return nil
}

// ToggleComments переключает возможность комментирования поста
func (s *Service) ToggleComments(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, enabled bool) (*model.Post, error) {
	s.logger.Debug("Toggling post comments",
		zap.String("post_id", postID.String()),
		zap.String("author_id", authorID.String()),
		zap.Bool("enabled", enabled),
	)

	input := model.PostUpdateInput{
		CommentsEnabled: &enabled,
	}

	updatedPost, err := s.UpdatePost(ctx, postID, input, authorID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Post comments toggled",
		zap.String("post_id", postID.String()),
		zap.Bool("comments_enabled", enabled),
	)

	return updatedPost, nil
}

// GetPostWithCommentCounts возвращает посты с количеством комментариев
func (s *Service) GetPostWithCommentCounts(ctx context.Context, filter model.PostFilter, pagination model.PaginationInput) ([]*model.Post, error) {
	s.logger.Debug("Getting posts with comment counts", zap.Any("filter", filter))

	repoFilter := converter.PostFilterToRepo(filter, pagination)

	repoPostsWithCounts, err := s.postRepo.ListWithCommentCounts(ctx, repoFilter)
	if err != nil {
		s.logger.Error("Failed to get posts with comment counts",
			zap.Error(err),
			zap.Any("filter", repoFilter),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to get posts with comment counts: %v", err))
	}

	posts := make([]*model.Post, len(repoPostsWithCounts))
	for i, repoPostWithCount := range repoPostsWithCounts {
		posts[i] = converter.PostFromRepo(&repoPostWithCount.Post)
		// Можно добавить информацию о количестве комментариев в будущем
	}

	s.logger.Debug("Posts with comment counts retrieved",
		zap.Int("count", len(posts)),
	)

	return posts, nil
}

// validatePagination проверяет корректность параметров пагинации
func (s *Service) validatePagination(pagination model.PaginationInput) error {
	const maxPageSize = 100

	if pagination.First != nil {
		if *pagination.First < 0 {
			return model.NewValidationError("first", "first must be non-negative")
		}
		if *pagination.First > maxPageSize {
			return model.NewValidationError("first", fmt.Sprintf("first cannot exceed %d", maxPageSize))
		}
	}

	if pagination.Last != nil {
		if *pagination.Last < 0 {
			return model.NewValidationError("last", "last must be non-negative")
		}
		if *pagination.Last > maxPageSize {
			return model.NewValidationError("last", fmt.Sprintf("last cannot exceed %d", maxPageSize))
		}
	}

	if pagination.First != nil && pagination.Last != nil {
		return model.NewValidationError("pagination", "cannot specify both first and last")
	}

	return nil
}

// buildPostConnection строит PostConnection с правильной пагинацией
func (s *Service) buildPostConnection(posts []*model.Post, pagination model.PaginationInput, totalCount int) *model.PostConnection {
	edges := make([]*model.PostEdge, len(posts))

	for i, post := range posts {
		// Создание курсора (base64 кодированный timestamp + ID для стабильной сортировки)
		cursorData := fmt.Sprintf("%d_%s", post.CreatedAt.Unix(), post.ID.String())
		cursor := base64.StdEncoding.EncodeToString([]byte(cursorData))

		edges[i] = &model.PostEdge{
			Node:   post,
			Cursor: cursor,
		}
	}

	// Определение информации о пагинации
	pageInfo := &model.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
	}

	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor

		// Более точное определение наличия следующей/предыдущей страницы
		if pagination.First != nil {
			requestedSize := *pagination.First
			pageInfo.HasNextPage = len(edges) == requestedSize && totalCount > requestedSize
		}

		if pagination.Last != nil {
			requestedSize := *pagination.Last
			pageInfo.HasPreviousPage = len(edges) == requestedSize && totalCount > requestedSize
		}
	}

	return &model.PostConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}
}

// decodeCursor декодирует курсор и возвращает timestamp и ID
func (s *Service) decodeCursor(cursor string) (time.Time, uuid.UUID, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor: %w", err)
	}

	parts := string(decoded)
	// Ожидаем формат "timestamp_uuid"
	timestampStr := parts[:strings.Index(parts, "_")]
	uuidStr := parts[strings.Index(parts, "_")+1:]

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor timestamp: %w", err)
	}

	id, err := uuid.Parse(uuidStr)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor UUID: %w", err)
	}

	return time.Unix(timestamp, 0), id, nil
}
