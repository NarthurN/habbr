package comment

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/NarthurN/habbr/internal/model"
	"github.com/NarthurN/habbr/internal/repository"
	"github.com/NarthurN/habbr/internal/repository/converter"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service реализует бизнес-логику для работы с комментариями
type Service struct {
	commentRepo     repository.CommentRepository
	postRepo        repository.PostRepository
	logger          *zap.Logger
	maxDepth        int
	subscriptionSvc SubscriptionNotifier
}

// SubscriptionNotifier интерфейс для уведомлений о комментариях
type SubscriptionNotifier interface {
	NotifyCommentCreated(ctx context.Context, comment *model.Comment) error
	NotifyCommentUpdated(ctx context.Context, comment *model.Comment) error
	NotifyCommentDeleted(ctx context.Context, postID uuid.UUID, commentID uuid.UUID) error
}

// NewService создает новый сервис комментариев
func NewService(repos *repository.Repositories, logger *zap.Logger, subscriptionSvc SubscriptionNotifier) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Service{
		commentRepo:     repos.Comment,
		postRepo:        repos.Post,
		logger:          logger,
		maxDepth:        50, // Ограничение глубины для предотвращения злоупотреблений
		subscriptionSvc: subscriptionSvc,
	}
}

// CreateComment создает новый комментарий
func (s *Service) CreateComment(ctx context.Context, input model.CommentInput) (*model.Comment, error) {
	s.logger.Debug("Creating new comment",
		zap.String("post_id", input.PostID.String()),
		zap.String("author_id", input.AuthorID.String()),
		zap.Bool("has_parent", input.ParentID != nil),
	)

	// Валидация входных данных
	if err := input.Validate(); err != nil {
		s.logger.Warn("Comment validation failed",
			zap.Error(err),
			zap.String("post_id", input.PostID.String()),
			zap.String("author_id", input.AuthorID.String()),
		)
		return nil, model.NewValidationError("input", err.Error())
	}

	// Проверка существования поста
	post, err := s.postRepo.GetByID(ctx, input.PostID)
	if err != nil {
		if err == repository.ErrNotFound {
			s.logger.Debug("Post not found for comment creation",
				zap.String("post_id", input.PostID.String()),
			)
			return nil, model.NewNotFoundError("post", input.PostID)
		}

		s.logger.Error("Failed to get post for comment creation",
			zap.Error(err),
			zap.String("post_id", input.PostID.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to get post: %v", err))
	}

	// Проверка возможности комментирования
	if !post.CommentsEnabled {
		s.logger.Warn("Attempt to comment on post with disabled comments",
			zap.String("post_id", input.PostID.String()),
			zap.String("author_id", input.AuthorID.String()),
		)
		return nil, model.NewForbiddenError("comments are disabled for this post")
	}

	// Определение глубины комментария
	depth := 0
	if input.ParentID != nil {
		parentComment, err := s.commentRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			if err == repository.ErrNotFound {
				s.logger.Debug("Parent comment not found",
					zap.String("parent_id", input.ParentID.String()),
				)
				return nil, model.NewNotFoundError("parent comment", *input.ParentID)
			}

			s.logger.Error("Failed to get parent comment",
				zap.Error(err),
				zap.String("parent_id", input.ParentID.String()),
			)
			return nil, model.NewInternalError(fmt.Sprintf("failed to get parent comment: %v", err))
		}

		// Проверка, что родительский комментарий принадлежит тому же посту
		if parentComment.PostID != input.PostID {
			s.logger.Warn("Parent comment belongs to different post",
				zap.String("parent_post_id", parentComment.PostID.String()),
				zap.String("expected_post_id", input.PostID.String()),
			)
			return nil, model.NewValidationError("parent_id", "parent comment must belong to the same post")
		}

		depth = parentComment.Depth + 1

		// Проверка максимальной глубины
		if depth > s.maxDepth {
			s.logger.Warn("Comment depth limit exceeded",
				zap.Int("depth", depth),
				zap.Int("max_depth", s.maxDepth),
				zap.String("post_id", input.PostID.String()),
			)
			return nil, model.NewValidationError("depth", fmt.Sprintf("comment depth cannot exceed %d", s.maxDepth))
		}
	}

	// Создание доменной модели
	comment := model.NewComment(input, depth)

	// Конвертация в модель репозитория и сохранение
	repoComment := converter.CommentToRepo(comment)
	if err := s.commentRepo.Create(ctx, repoComment); err != nil {
		s.logger.Error("Failed to create comment in repository",
			zap.Error(err),
			zap.String("comment_id", comment.ID.String()),
			zap.String("post_id", comment.PostID.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to create comment: %v", err))
	}

	s.logger.Info("Comment created successfully",
		zap.String("comment_id", comment.ID.String()),
		zap.String("post_id", comment.PostID.String()),
		zap.String("author_id", comment.AuthorID.String()),
		zap.Int("depth", comment.Depth),
	)

	// Отправляем уведомление о новом комментарии
	if s.subscriptionSvc != nil {
		if err := s.subscriptionSvc.NotifyCommentCreated(ctx, comment); err != nil {
			s.logger.Warn("Failed to send comment creation notification",
				zap.Error(err),
				zap.String("comment_id", comment.ID.String()),
			)
			// Не прерываем выполнение, уведомления не критичны
		}
	}

	return comment, nil
}

// GetComment возвращает комментарий по ID
func (s *Service) GetComment(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	if id == uuid.Nil {
		s.logger.Warn("Attempt to get comment with nil ID")
		return nil, model.NewValidationError("id", "comment ID is required")
	}

	s.logger.Debug("Getting comment by ID", zap.String("comment_id", id.String()))

	repoComment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			s.logger.Debug("Comment not found", zap.String("comment_id", id.String()))
			return nil, model.NewNotFoundError("comment", id)
		}

		s.logger.Error("Failed to get comment from repository",
			zap.Error(err),
			zap.String("comment_id", id.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to get comment: %v", err))
	}

	comment := converter.CommentFromRepo(repoComment)
	s.logger.Debug("Comment retrieved successfully",
		zap.String("comment_id", comment.ID.String()),
		zap.String("post_id", comment.PostID.String()),
	)

	return comment, nil
}

// ListComments возвращает список комментариев с пагинацией
func (s *Service) ListComments(ctx context.Context, filter model.CommentFilter, pagination model.PaginationInput) (*model.CommentConnection, error) {
	s.logger.Debug("Listing comments",
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
	)

	// Валидация пагинации
	if err := s.validatePagination(pagination); err != nil {
		s.logger.Warn("Invalid pagination parameters", zap.Error(err))
		return nil, err
	}

	// Конвертация фильтра
	repoFilter := converter.CommentFilterToRepo(filter, pagination)

	// Получение комментариев
	repoComments, err := s.commentRepo.List(ctx, repoFilter)
	if err != nil {
		s.logger.Error("Failed to list comments from repository",
			zap.Error(err),
			zap.Any("filter", repoFilter),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to list comments: %v", err))
	}

	// Получение общего количества
	totalCount, err := s.commentRepo.Count(ctx, repoFilter)
	if err != nil {
		s.logger.Error("Failed to count comments from repository",
			zap.Error(err),
			zap.Any("filter", repoFilter),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to count comments: %v", err))
	}

	// Конвертация в доменные модели
	comments := converter.CommentsFromRepo(repoComments)

	// Создание connection с пагинацией
	connection := s.buildCommentConnection(comments, pagination, totalCount)

	s.logger.Debug("Comments listed successfully",
		zap.Int("count", len(comments)),
		zap.Int("total_count", totalCount),
		zap.Bool("has_next_page", connection.PageInfo.HasNextPage),
	)

	return connection, nil
}

// GetCommentsTree возвращает дерево комментариев к посту
func (s *Service) GetCommentsTree(ctx context.Context, postID uuid.UUID) ([]*model.Comment, error) {
	if postID == uuid.Nil {
		s.logger.Warn("Attempt to get comments tree with nil post ID")
		return nil, model.NewValidationError("post_id", "post ID is required")
	}

	s.logger.Debug("Getting comments tree", zap.String("post_id", postID.String()))

	// Проверка существования поста
	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		s.logger.Error("Failed to check post existence",
			zap.Error(err),
			zap.String("post_id", postID.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to check post existence: %v", err))
	}

	if !exists {
		s.logger.Debug("Post not found for comments tree",
			zap.String("post_id", postID.String()),
		)
		return nil, model.NewNotFoundError("post", postID)
	}

	// Получение всех комментариев к посту
	repoComments, err := s.commentRepo.GetByPostID(ctx, postID)
	if err != nil {
		s.logger.Error("Failed to get post comments from repository",
			zap.Error(err),
			zap.String("post_id", postID.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to get post comments: %v", err))
	}

	// Конвертация в доменные модели
	comments := converter.CommentsFromRepo(repoComments)

	// Построение дерева
	tree := model.BuildCommentsTree(comments)

	s.logger.Debug("Comments tree built successfully",
		zap.String("post_id", postID.String()),
		zap.Int("total_comments", len(comments)),
		zap.Int("root_comments", len(tree)),
	)

	return tree, nil
}

// UpdateComment обновляет комментарий
func (s *Service) UpdateComment(ctx context.Context, id uuid.UUID, input model.CommentUpdateInput, authorID uuid.UUID) (*model.Comment, error) {
	s.logger.Debug("Updating comment",
		zap.String("comment_id", id.String()),
		zap.String("author_id", authorID.String()),
	)

	// Валидация входных данных
	if err := input.Validate(); err != nil {
		s.logger.Warn("Comment update validation failed",
			zap.Error(err),
			zap.String("comment_id", id.String()),
		)
		return nil, model.NewValidationError("input", err.Error())
	}

	if id == uuid.Nil {
		s.logger.Warn("Attempt to update comment with nil ID")
		return nil, model.NewValidationError("id", "comment ID is required")
	}

	if authorID == uuid.Nil {
		s.logger.Warn("Attempt to update comment with nil author ID",
			zap.String("comment_id", id.String()),
		)
		return nil, model.NewValidationError("author_id", "author ID is required")
	}

	// Получение существующего комментария
	existingComment, err := s.GetComment(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверка прав на редактирование
	if existingComment.AuthorID != authorID {
		s.logger.Warn("Unauthorized attempt to update comment",
			zap.String("comment_id", id.String()),
			zap.String("comment_author", existingComment.AuthorID.String()),
			zap.String("requesting_user", authorID.String()),
		)
		return nil, model.NewForbiddenError("update comment")
	}

	// Сохранение исходного контента для логирования
	originalContent := existingComment.Content

	// Обновление комментария
	existingComment.Update(input)

	// Сохранение изменений
	repoComment := converter.CommentToRepo(existingComment)
	if err := s.commentRepo.Update(ctx, repoComment); err != nil {
		if err == repository.ErrNotFound {
			return nil, model.NewNotFoundError("comment", id)
		}

		s.logger.Error("Failed to update comment in repository",
			zap.Error(err),
			zap.String("comment_id", id.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to update comment: %v", err))
	}

	s.logger.Info("Comment updated successfully",
		zap.String("comment_id", id.String()),
		zap.String("author_id", authorID.String()),
		zap.String("old_content", originalContent),
		zap.String("new_content", existingComment.Content),
	)

	// Отправляем уведомление об обновлении комментария
	if s.subscriptionSvc != nil {
		if err := s.subscriptionSvc.NotifyCommentUpdated(ctx, existingComment); err != nil {
			s.logger.Warn("Failed to send comment update notification",
				zap.Error(err),
				zap.String("comment_id", existingComment.ID.String()),
			)
		}
	}

	return existingComment, nil
}

// DeleteComment удаляет комментарий
func (s *Service) DeleteComment(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error {
	s.logger.Debug("Deleting comment",
		zap.String("comment_id", id.String()),
		zap.String("author_id", authorID.String()),
	)

	if id == uuid.Nil {
		s.logger.Warn("Attempt to delete comment with nil ID")
		return model.NewValidationError("id", "comment ID is required")
	}

	if authorID == uuid.Nil {
		s.logger.Warn("Attempt to delete comment with nil author ID",
			zap.String("comment_id", id.String()),
		)
		return model.NewValidationError("author_id", "author ID is required")
	}

	// Получение комментария для проверки прав
	comment, err := s.GetComment(ctx, id)
	if err != nil {
		return err
	}

	// Проверка прав на удаление
	if comment.AuthorID != authorID {
		s.logger.Warn("Unauthorized attempt to delete comment",
			zap.String("comment_id", id.String()),
			zap.String("comment_author", comment.AuthorID.String()),
			zap.String("requesting_user", authorID.String()),
		)
		return model.NewForbiddenError("delete comment")
	}

	// Получение дочерних комментариев для подсчета
	children, err := s.commentRepo.GetChildren(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get child comments for deletion",
			zap.Error(err),
			zap.String("comment_id", id.String()),
		)
		return model.NewInternalError(fmt.Sprintf("failed to get child comments: %v", err))
	}

	deletedCount := len(children) + 1 // включая сам комментарий

	// Удаление всех дочерних комментариев (каскадное удаление)
	for _, child := range children {
		if err := s.commentRepo.Delete(ctx, child.ID); err != nil {
			s.logger.Error("Failed to delete child comment",
				zap.Error(err),
				zap.String("child_comment_id", child.ID.String()),
				zap.String("parent_comment_id", id.String()),
			)
			return model.NewInternalError(fmt.Sprintf("failed to delete child comment: %v", err))
		}
	}

	// Удаление самого комментария
	if err := s.commentRepo.Delete(ctx, id); err != nil {
		if err == repository.ErrNotFound {
			return model.NewNotFoundError("comment", id)
		}

		s.logger.Error("Failed to delete comment from repository",
			zap.Error(err),
			zap.String("comment_id", id.String()),
		)
		return model.NewInternalError(fmt.Sprintf("failed to delete comment: %v", err))
	}

	s.logger.Info("Comment deleted successfully",
		zap.String("comment_id", id.String()),
		zap.String("post_id", comment.PostID.String()),
		zap.String("author_id", authorID.String()),
		zap.Int("deleted_comments_count", deletedCount),
	)

	// Отправляем уведомление об удалении комментария
	if s.subscriptionSvc != nil {
		if err := s.subscriptionSvc.NotifyCommentDeleted(ctx, comment.PostID, id); err != nil {
			s.logger.Warn("Failed to send comment deletion notification",
				zap.Error(err),
				zap.String("comment_id", id.String()),
			)
		}
	}

	return nil
}

// GetCommentWithChildren возвращает комментарий со всеми дочерними комментариями
func (s *Service) GetCommentWithChildren(ctx context.Context, commentID uuid.UUID) (*model.Comment, error) {
	s.logger.Debug("Getting comment with children", zap.String("comment_id", commentID.String()))

	// Получение основного комментария
	comment, err := s.GetComment(ctx, commentID)
	if err != nil {
		return nil, err
	}

	// Получение всех комментариев к посту для построения дерева
	repoComments, err := s.commentRepo.GetByPostID(ctx, comment.PostID)
	if err != nil {
		s.logger.Error("Failed to get post comments for building tree",
			zap.Error(err),
			zap.String("post_id", comment.PostID.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to get post comments: %v", err))
	}

	// Конвертация в доменные модели и построение дерева
	allComments := converter.CommentsFromRepo(repoComments)
	tree := model.BuildCommentsTree(allComments)

	// Поиск нужного комментария в дереве
	targetComment := s.findCommentInTree(tree, commentID)
	if targetComment == nil {
		return comment, nil // Возвращаем комментарий без детей, если не нашли в дереве
	}

	return targetComment, nil
}

// GetCommentDepthStatistics возвращает статистику глубины комментариев для поста
func (s *Service) GetCommentDepthStatistics(ctx context.Context, postID uuid.UUID) (map[int]int, error) {
	s.logger.Debug("Getting comment depth statistics", zap.String("post_id", postID.String()))

	repoComments, err := s.commentRepo.GetByPostID(ctx, postID)
	if err != nil {
		s.logger.Error("Failed to get comments for depth statistics",
			zap.Error(err),
			zap.String("post_id", postID.String()),
		)
		return nil, model.NewInternalError(fmt.Sprintf("failed to get comments: %v", err))
	}

	statistics := make(map[int]int)
	for _, comment := range repoComments {
		statistics[comment.Depth]++
	}

	s.logger.Debug("Comment depth statistics calculated",
		zap.String("post_id", postID.String()),
		zap.Any("statistics", statistics),
	)

	return statistics, nil
}

// findCommentInTree рекурсивно ищет комментарий в дереве
func (s *Service) findCommentInTree(tree []*model.Comment, commentID uuid.UUID) *model.Comment {
	for _, comment := range tree {
		if comment.ID == commentID {
			return comment
		}
		if len(comment.Children) > 0 {
			if found := s.findCommentInTree(comment.Children, commentID); found != nil {
				return found
			}
		}
	}
	return nil
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

// buildCommentConnection строит CommentConnection с правильной пагинацией
func (s *Service) buildCommentConnection(comments []*model.Comment, pagination model.PaginationInput, totalCount int) *model.CommentConnection {
	edges := make([]*model.CommentEdge, len(comments))

	for i, comment := range comments {
		// Создание курсора (base64 кодированный timestamp + ID для стабильной сортировки)
		cursorData := fmt.Sprintf("%d_%s", comment.CreatedAt.Unix(), comment.ID.String())
		cursor := base64.StdEncoding.EncodeToString([]byte(cursorData))

		edges[i] = &model.CommentEdge{
			Node:   comment,
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

	return &model.CommentConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}
}
