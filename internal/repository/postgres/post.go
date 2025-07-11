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

// PostRepository реализует repository.PostRepository для PostgreSQL
type PostRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewPostRepository создает новый PostgreSQL репозиторий постов
func NewPostRepository(pool *pgxpool.Pool, logger *zap.Logger) repository.PostRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PostRepository{
		pool:   pool,
		logger: logger,
	}
}

// Create создает новый пост в базе данных
func (r *PostRepository) Create(ctx context.Context, post *repomodel.Post) error {
	if post == nil {
		return fmt.Errorf("post cannot be nil")
	}

	query := `
		INSERT INTO posts (id, title, content, author_id, comments_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		post.ID,
		post.Title,
		post.Content,
		post.AuthorID,
		post.CommentsEnabled,
		post.CreatedAt,
		post.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create post",
			zap.String("post_id", post.ID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create post: %w", err)
	}

	r.logger.Debug("Post created successfully", zap.String("post_id", post.ID.String()))
	return nil
}

// GetByID получает пост по ID
func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*repomodel.Post, error) {
	query := `
		SELECT id, title, content, author_id, comments_enabled, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var post repomodel.Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.CommentsEnabled,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		r.logger.Error("Failed to get post by ID",
			zap.String("post_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return &post, nil
}

// List получает список постов с фильтрацией и пагинацией
func (r *PostRepository) List(ctx context.Context, filter repomodel.PostFilter) ([]*repomodel.Post, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT id, title, content, author_id, comments_enabled, created_at, updated_at
		FROM posts
	`

	// Добавляем условия фильтрации
	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argIndex))
		args = append(args, *filter.AuthorID)
		argIndex++
	}

	if filter.WithComments != nil {
		conditions = append(conditions, fmt.Sprintf("comments_enabled = $%d", argIndex))
		args = append(args, *filter.WithComments)
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
	orderDir := "DESC"
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
		r.logger.Error("Failed to list posts", zap.Error(err))
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}
	defer rows.Close()

	var posts []*repomodel.Post
	for rows.Next() {
		var post repomodel.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.CommentsEnabled,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan post", zap.Error(err))
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating posts", zap.Error(err))
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, nil
}

// Count подсчитывает общее количество постов с фильтрацией
func (r *PostRepository) Count(ctx context.Context, filter repomodel.PostFilter) (int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := "SELECT COUNT(*) FROM posts"

	// Добавляем условия фильтрации
	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argIndex))
		args = append(args, *filter.AuthorID)
		argIndex++
	}

	if filter.WithComments != nil {
		conditions = append(conditions, fmt.Sprintf("comments_enabled = $%d", argIndex))
		args = append(args, *filter.WithComments)
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count posts", zap.Error(err))
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}

	return count, nil
}

// Update обновляет пост
func (r *PostRepository) Update(ctx context.Context, post *repomodel.Post) error {
	if post == nil {
		return fmt.Errorf("post cannot be nil")
	}

	query := `
		UPDATE posts
		SET title = $2, content = $3, comments_enabled = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		post.ID,
		post.Title,
		post.Content,
		post.CommentsEnabled,
		post.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to update post",
			zap.String("post_id", post.ID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	r.logger.Debug("Post updated successfully", zap.String("post_id", post.ID.String()))
	return nil
}

// Delete удаляет пост
func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM posts WHERE id = $1"

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete post",
			zap.String("post_id", id.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	r.logger.Debug("Post deleted successfully", zap.String("post_id", id.String()))
	return nil
}

// Exists проверяет существование поста
func (r *PostRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)"

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check post existence",
			zap.String("post_id", id.String()),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to check post existence: %w", err)
	}

	return exists, nil
}

// ListWithCommentCounts получает посты с количеством комментариев
func (r *PostRepository) ListWithCommentCounts(ctx context.Context, filter repomodel.PostFilter) ([]*repomodel.PostWithCommentCount, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT
			p.id, p.title, p.content, p.author_id, p.comments_enabled,
			p.created_at, p.updated_at,
			COALESCE(c.comment_count, 0) as comment_count
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as comment_count
			FROM comments
			GROUP BY post_id
		) c ON p.id = c.post_id
	`

	// Добавляем условия фильтрации
	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("p.author_id = $%d", argIndex))
		args = append(args, *filter.AuthorID)
		argIndex++
	}

	if filter.WithComments != nil {
		conditions = append(conditions, fmt.Sprintf("p.comments_enabled = $%d", argIndex))
		args = append(args, *filter.WithComments)
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
		orderBy = "p." + filter.OrderBy
	}
	orderDir := "DESC"
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
		r.logger.Error("Failed to list posts with comment counts", zap.Error(err))
		return nil, fmt.Errorf("failed to list posts with comment counts: %w", err)
	}
	defer rows.Close()

	var posts []*repomodel.PostWithCommentCount
	for rows.Next() {
		var postWithCount repomodel.PostWithCommentCount
		err := rows.Scan(
			&postWithCount.Post.ID,
			&postWithCount.Post.Title,
			&postWithCount.Post.Content,
			&postWithCount.Post.AuthorID,
			&postWithCount.Post.CommentsEnabled,
			&postWithCount.Post.CreatedAt,
			&postWithCount.Post.UpdatedAt,
			&postWithCount.CommentCount,
		)
		if err != nil {
			r.logger.Error("Failed to scan post with comment count", zap.Error(err))
			return nil, fmt.Errorf("failed to scan post with comment count: %w", err)
		}
		posts = append(posts, &postWithCount)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating posts with comment counts", zap.Error(err))
		return nil, fmt.Errorf("error iterating posts with comment counts: %w", err)
	}

	return posts, nil
}
