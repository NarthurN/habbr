//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/NarthurN/habbr/internal/config"
	repomodel "github.com/NarthurN/habbr/internal/repository/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestPostRepository_Integration тестирует PostgreSQL репозиторий постов
func TestPostRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	// Настройка тестовой базы данных (требует запущенный PostgreSQL)
	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5432,
		Name:           "habbr_test",
		User:           "postgres",
		Password:       "password",
		SSLMode:        "disable",
		MaxConnections: 5,
		MaxIdleTime:    time.Minute,
		MaxLifetime:    time.Hour,
	}

	// Создаем менеджер
	manager, err := NewManager(ctx, cfg, logger)
	require.NoError(t, err)
	defer manager.Close(ctx)

	// Выполняем миграции
	err = manager.Migrate(ctx)
	require.NoError(t, err)

	repo := manager.GetRepositories().Post

	t.Run("Create and Get Post", func(t *testing.T) {
		// Создаем тестовый пост
		post := &repomodel.Post{
			ID:              uuid.New(),
			Title:           "Test Post",
			Content:         "This is a test post content",
			AuthorID:        uuid.New(),
			CommentsEnabled: true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Создаем пост
		err := repo.Create(ctx, post)
		require.NoError(t, err)

		// Получаем пост
		retrieved, err := repo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, post.ID, retrieved.ID)
		assert.Equal(t, post.Title, retrieved.Title)
		assert.Equal(t, post.Content, retrieved.Content)
		assert.Equal(t, post.AuthorID, retrieved.AuthorID)
		assert.Equal(t, post.CommentsEnabled, retrieved.CommentsEnabled)

		// Проверяем существование
		exists, err := repo.Exists(ctx, post.ID)
		require.NoError(t, err)
		assert.True(t, exists)

		// Обновляем пост
		post.Title = "Updated Title"
		post.Content = "Updated content"
		post.UpdatedAt = time.Now()
		err = repo.Update(ctx, post)
		require.NoError(t, err)

		// Проверяем обновление
		updated, err := repo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "Updated content", updated.Content)

		// Удаляем пост
		err = repo.Delete(ctx, post.ID)
		require.NoError(t, err)

		// Проверяем удаление
		exists, err = repo.Exists(ctx, post.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("List and Count Posts", func(t *testing.T) {
		authorID := uuid.New()

		// Создаем несколько постов
		posts := []*repomodel.Post{
			{
				ID:              uuid.New(),
				Title:           "Post 1",
				Content:         "Content 1",
				AuthorID:        authorID,
				CommentsEnabled: true,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				ID:              uuid.New(),
				Title:           "Post 2",
				Content:         "Content 2",
				AuthorID:        authorID,
				CommentsEnabled: false,
				CreatedAt:       time.Now().Add(time.Minute),
				UpdatedAt:       time.Now().Add(time.Minute),
			},
		}

		for _, post := range posts {
			err := repo.Create(ctx, post)
			require.NoError(t, err)
		}

		// Тестируем список без фильтров
		filter := repomodel.PostFilter{
			Limit: 10,
		}
		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)

		// Тестируем фильтр по автору
		filter = repomodel.PostFilter{
			AuthorID: &authorID,
			Limit:    10,
		}
		result, err = repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Тестируем подсчет
		count, err := repo.Count(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		// Очищаем тестовые данные
		for _, post := range posts {
			repo.Delete(ctx, post.ID)
		}
	})
}

// TestCommentRepository_Integration тестирует PostgreSQL репозиторий комментариев
func TestCommentRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	// Настройка тестовой базы данных
	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5432,
		Name:           "habbr_test",
		User:           "postgres",
		Password:       "password",
		SSLMode:        "disable",
		MaxConnections: 5,
		MaxIdleTime:    time.Minute,
		MaxLifetime:    time.Hour,
	}

	manager, err := NewManager(ctx, cfg, logger)
	require.NoError(t, err)
	defer manager.Close(ctx)

	err = manager.Migrate(ctx)
	require.NoError(t, err)

	postRepo := manager.GetRepositories().Post
	commentRepo := manager.GetRepositories().Comment

	t.Run("Hierarchical Comments", func(t *testing.T) {
		// Создаем тестовый пост
		post := &repomodel.Post{
			ID:              uuid.New(),
			Title:           "Test Post for Comments",
			Content:         "Content",
			AuthorID:        uuid.New(),
			CommentsEnabled: true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		err := postRepo.Create(ctx, post)
		require.NoError(t, err)

		// Создаем корневой комментарий
		rootComment := &repomodel.Comment{
			ID:        uuid.New(),
			PostID:    post.ID,
			ParentID:  nil,
			Content:   "Root comment",
			AuthorID:  uuid.New(),
			Depth:     0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = commentRepo.Create(ctx, rootComment)
		require.NoError(t, err)

		// Создаем дочерний комментарий
		childComment := &repomodel.Comment{
			ID:        uuid.New(),
			PostID:    post.ID,
			ParentID:  &rootComment.ID,
			Content:   "Child comment",
			AuthorID:  uuid.New(),
			Depth:     1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = commentRepo.Create(ctx, childComment)
		require.NoError(t, err)

		// Получаем все комментарии к посту
		comments, err := commentRepo.GetByPostID(ctx, post.ID)
		require.NoError(t, err)
		assert.Len(t, comments, 2)

		// Проверяем сортировку по depth
		assert.Equal(t, 0, comments[0].Depth)
		assert.Equal(t, 1, comments[1].Depth)

		// Получаем дочерние комментарии
		children, err := commentRepo.GetChildren(ctx, rootComment.ID)
		require.NoError(t, err)
		assert.Len(t, children, 1)
		assert.Equal(t, childComment.ID, children[0].ID)

		// Получаем максимальную глубину
		maxDepth, err := commentRepo.GetMaxDepthForPost(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, maxDepth)

		// Подсчитываем комментарии к посту
		count, err := commentRepo.CountByPostID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		// Удаляем все комментарии к посту
		err = commentRepo.DeleteByPostID(ctx, post.ID)
		require.NoError(t, err)

		// Проверяем удаление
		count, err = commentRepo.CountByPostID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Очищаем тестовый пост
		postRepo.Delete(ctx, post.ID)
	})
}

// TestManager_Migration тестирует систему миграций
func TestManager_Migration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	cfg := &config.DatabaseConfig{
		Host:           "localhost",
		Port:           5432,
		Name:           "habbr_test",
		User:           "postgres",
		Password:       "password",
		SSLMode:        "disable",
		MaxConnections: 5,
		MaxIdleTime:    time.Minute,
		MaxLifetime:    time.Hour,
	}

	manager, err := NewManager(ctx, cfg, logger)
	require.NoError(t, err)
	defer manager.Close(ctx)

	// Выполняем миграции дважды - вторая должна быть идемпотентной
	err = manager.Migrate(ctx)
	require.NoError(t, err)

	err = manager.Migrate(ctx)
	require.NoError(t, err)

	// Проверяем health check
	err = manager.HealthCheck(ctx)
	require.NoError(t, err)
}
