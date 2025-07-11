package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/NarthurN/habbr/internal/repository"
	repomodel "github.com/NarthurN/habbr/internal/repository/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// CommentRepository реализует repository.CommentRepository для PostgreSQL
type CommentRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewCommentRepository создает новый PostgreSQL репозиторий комментариев
func NewCommentRepository(pool *pgxpool.Pool, logger *zap.Logger) repository.CommentRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CommentRepository{
		pool:   pool,
		logger: logger,
	}
}

// Create создает новый комментарий в базе данных
func (r *CommentRepository) Create(ctx context.Context, comment *repomodel.Comment) error {
	if comment == nil {
		return fmt.Errorf("comment cannot be nil")
	}

	query := `
		INSERT INTO comments (id, post_id, parent_id, content, author_id, depth, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		comment.ID,
		comment.PostID,
		comment.ParentID,
		comment.Content,
		comment.AuthorID,
		comment.Depth,
		comment.CreatedAt,
		comment.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create comment",
			zap.String("comment_id", comment.ID.String()),
			zap.String("post_id", comment.PostID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create comment: %w", err)
	}

	r.logger.Debug("Comment created successfully",
		zap.String("comment_id", comment.ID.String()),
		zap.Int("depth", comment.Depth),
	)
	return nil
}

// GetByID получает комментарий по ID
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*repomodel.Comment, error) {
	query := `
		SELECT id, post_id, parent_id, content, author_id, depth, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var comment repomodel.Comment
	err := row.Scan(
		&comment.ID,
		&comment.PostID,
		&comment.ParentID,
		&comment.Content,
		&comment.AuthorID,
		&comment.Depth,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		r.logger.Error("Failed to get comment by ID",
			zap.String("comment_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &comment, nil
}

// List получает список комментариев с фильтрацией и пагинацией
func (r *CommentRepository) List(ctx context.Context, filter repomodel.CommentFilter) ([]*repomodel.Comment, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT id, post_id, parent_id, content, author_id, depth, created_at, updated_at
		FROM comments
	`

	// Добавляем условия фильтрации
	if filter.PostID != nil {
		conditions = append(conditions, fmt.Sprintf("post_id = $%d", argIndex))
		args = append(args, *filter.PostID)
		argIndex++
	}

	if filter.ParentID != nil {
		conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argIndex))
		args = append(args, *filter.ParentID)
		argIndex++
	}

	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argIndex))
		args = append(args, *filter.AuthorID)
		argIndex++
	}

	if filter.MaxDepth != nil {
		conditions = append(conditions, fmt.Sprintf("depth <= $%d", argIndex))
		args = append(args, *filter.MaxDepth)
		argIndex++
	}

	// Собираем полный запрос
	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Добавляем сортировку
	orderBy := "created_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDir := "ASC"
	if filter.OrderDir != "" {
		orderDir = strings.ToUpper(filter.OrderDir)
	}
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// Добавляем пагинацию
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list comments", zap.Error(err))
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []*repomodel.Comment
	for rows.Next() {
		var comment repomodel.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Content,
			&comment.AuthorID,
			&comment.Depth,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan comment", zap.Error(err))
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating comments", zap.Error(err))
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// Count подсчитывает общее количество комментариев с фильтрацией
func (r *CommentRepository) Count(ctx context.Context, filter repomodel.CommentFilter) (int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := "SELECT COUNT(*) FROM comments"

	// Добавляем условия фильтрации
	if filter.PostID != nil {
		conditions = append(conditions, fmt.Sprintf("post_id = $%d", argIndex))
		args = append(args, *filter.PostID)
		argIndex++
	}

	if filter.ParentID != nil {
		conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argIndex))
		args = append(args, *filter.ParentID)
		argIndex++
	}

	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argIndex))
		args = append(args, *filter.AuthorID)
		argIndex++
	}

	if filter.MaxDepth != nil {
		conditions = append(conditions, fmt.Sprintf("depth <= $%d", argIndex))
		args = append(args, *filter.MaxDepth)
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count comments", zap.Error(err))
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}

	return count, nil
}

// Update обновляет комментарий
func (r *CommentRepository) Update(ctx context.Context, comment *repomodel.Comment) error {
	if comment == nil {
		return fmt.Errorf("comment cannot be nil")
	}

	query := `
		UPDATE comments
		SET content = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		comment.ID,
		comment.Content,
		comment.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to update comment",
			zap.String("comment_id", comment.ID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	r.logger.Debug("Comment updated successfully", zap.String("comment_id", comment.ID.String()))
	return nil
}

// Delete удаляет комментарий
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM comments WHERE id = $1"

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete comment",
			zap.String("comment_id", id.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	r.logger.Debug("Comment deleted successfully", zap.String("comment_id", id.String()))
	return nil
}

// Exists проверяет существование комментария
func (r *CommentRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)"

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check comment existence",
			zap.String("comment_id", id.String()),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to check comment existence: %w", err)
	}

	return exists, nil
}

// GetByPostID получает все комментарии к посту (для построения дерева)
func (r *CommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID) ([]*repomodel.Comment, error) {
	query := `
		SELECT id, post_id, parent_id, content, author_id, depth, created_at, updated_at
		FROM comments
		WHERE post_id = $1
		ORDER BY depth ASC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, postID)
	if err != nil {
		r.logger.Error("Failed to get comments by post ID",
			zap.String("post_id", postID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get comments by post ID: %w", err)
	}
	defer rows.Close()

	var comments []*repomodel.Comment
	for rows.Next() {
		var comment repomodel.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Content,
			&comment.AuthorID,
			&comment.Depth,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan comment", zap.Error(err))
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating comments", zap.Error(err))
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// GetChildren получает дочерние комментарии
func (r *CommentRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*repomodel.Comment, error) {
	query := `
		SELECT id, post_id, parent_id, content, author_id, depth, created_at, updated_at
		FROM comments
		WHERE parent_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		r.logger.Error("Failed to get child comments",
			zap.String("parent_id", parentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get child comments: %w", err)
	}
	defer rows.Close()

	var comments []*repomodel.Comment
	for rows.Next() {
		var comment repomodel.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Content,
			&comment.AuthorID,
			&comment.Depth,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan child comment", zap.Error(err))
			return nil, fmt.Errorf("failed to scan child comment: %w", err)
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating child comments", zap.Error(err))
		return nil, fmt.Errorf("error iterating child comments: %w", err)
	}

	return comments, nil
}

// GetMaxDepthForPost получает максимальную глубину комментариев для поста
func (r *CommentRepository) GetMaxDepthForPost(ctx context.Context, postID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(MAX(depth), 0)
		FROM comments
		WHERE post_id = $1
	`

	var maxDepth int
	err := r.pool.QueryRow(ctx, query, postID).Scan(&maxDepth)
	if err != nil {
		r.logger.Error("Failed to get max depth for post",
			zap.String("post_id", postID.String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to get max depth for post: %w", err)
	}

	return maxDepth, nil
}

// DeleteByPostID удаляет все комментарии к посту
func (r *CommentRepository) DeleteByPostID(ctx context.Context, postID uuid.UUID) error {
	query := "DELETE FROM comments WHERE post_id = $1"

	result, err := r.pool.Exec(ctx, query, postID)
	if err != nil {
		r.logger.Error("Failed to delete comments by post ID",
			zap.String("post_id", postID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete comments by post ID: %w", err)
	}

	rowsAffected := result.RowsAffected()
	r.logger.Debug("Comments deleted by post ID",
		zap.String("post_id", postID.String()),
		zap.Int64("rows_affected", rowsAffected),
	)

	return nil
}

// CountByPostID получает количество комментариев к посту
func (r *CommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int, error) {
	query := "SELECT COUNT(*) FROM comments WHERE post_id = $1"

	var count int
	err := r.pool.QueryRow(ctx, query, postID).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count comments by post ID",
			zap.String("post_id", postID.String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to count comments by post ID: %w", err)
	}

	return count, nil
}

// GetCommentsWithPagination получает комментарии с курсорной пагинацией
func (r *CommentRepository) GetCommentsWithPagination(ctx context.Context, postID uuid.UUID, cursor *string, limit int) ([]*repomodel.Comment, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT id, post_id, parent_id, content, author_id, depth, created_at, updated_at
		FROM comments
	`

	// Основной фильтр по посту
	conditions = append(conditions, fmt.Sprintf("post_id = $%d", argIndex))
	args = append(args, postID)
	argIndex++

	// Курсорная пагинация
	if cursor != nil && *cursor != "" {
		// Курсор - это timestamp в RFC3339 формате
		conditions = append(conditions, fmt.Sprintf("created_at > $%d", argIndex))
		args = append(args, *cursor)
		argIndex++
	}

	query := baseQuery + " WHERE " + strings.Join(conditions, " AND ")
	query += " ORDER BY depth ASC, created_at ASC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get comments with pagination",
			zap.String("post_id", postID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get comments with pagination: %w", err)
	}
	defer rows.Close()

	var comments []*repomodel.Comment
	for rows.Next() {
		var comment repomodel.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Content,
			&comment.AuthorID,
			&comment.Depth,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan comment", zap.Error(err))
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating comments", zap.Error(err))
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// GetCommentPath получает путь от корневого комментария до указанного
func (r *CommentRepository) GetCommentPath(ctx context.Context, commentID uuid.UUID) ([]*repomodel.Comment, error) {
	query := `
		WITH RECURSIVE comment_path AS (
			-- Базовый случай: начинаем с указанного комментария
			SELECT id, post_id, parent_id, content, author_id, depth, created_at, updated_at, 0 as level
			FROM comments
			WHERE id = $1

			UNION ALL

			-- Рекурсивный случай: поднимаемся к родителям
			SELECT c.id, c.post_id, c.parent_id, c.content, c.author_id, c.depth, c.created_at, c.updated_at, cp.level + 1
			FROM comments c
			INNER JOIN comment_path cp ON c.id = cp.parent_id
		)
		SELECT id, post_id, parent_id, content, author_id, depth, created_at, updated_at
		FROM comment_path
		ORDER BY level DESC
	`

	rows, err := r.pool.Query(ctx, query, commentID)
	if err != nil {
		r.logger.Error("Failed to get comment path",
			zap.String("comment_id", commentID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get comment path: %w", err)
	}
	defer rows.Close()

	var comments []*repomodel.Comment
	for rows.Next() {
		var comment repomodel.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Content,
			&comment.AuthorID,
			&comment.Depth,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan comment in path", zap.Error(err))
			return nil, fmt.Errorf("failed to scan comment in path: %w", err)
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating comment path", zap.Error(err))
		return nil, fmt.Errorf("error iterating comment path: %w", err)
	}

	return comments, nil
}
