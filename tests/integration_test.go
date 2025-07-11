//go:build integration

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/NarthurN/habbr/internal/config"
	"github.com/NarthurN/habbr/internal/model"
	"github.com/NarthurN/habbr/internal/repository"
	"github.com/NarthurN/habbr/internal/repository/memory"
	"github.com/NarthurN/habbr/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestPostService_Integration(t *testing.T) {
	// Setup
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "memory",
		},
	}

	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository()

	repos := &repository.Repositories{
		Post:    postRepo,
		Comment: commentRepo,
	}

	serviceManager := service.NewManager(repos, logger)
	services := serviceManager.GetServices()
	defer serviceManager.Close()

	postService := services.Post
	commentService := services.Comment

	ctx := context.Background()

	t.Run("full post lifecycle", func(t *testing.T) {
		// Create post
		authorID := uuid.New()
		postInput := &model.PostInput{
			Title:           "Integration Test Post",
			Content:         "This is a test post for integration testing",
			AuthorID:        authorID,
			CommentsEnabled: true,
		}

		createdPost, err := postService.CreatePost(ctx, postInput)
		require.NoError(t, err)
		require.NotNil(t, createdPost)
		assert.Equal(t, postInput.Title, createdPost.Title)
		assert.Equal(t, postInput.Content, createdPost.Content)
		assert.Equal(t, postInput.AuthorID, createdPost.AuthorID)
		assert.True(t, createdPost.CommentsEnabled)
		assert.NotEmpty(t, createdPost.ID)
		assert.WithinDuration(t, time.Now(), createdPost.CreatedAt, time.Second)

		// Get post
		retrievedPost, err := postService.GetPost(ctx, createdPost.ID)
		require.NoError(t, err)
		assert.Equal(t, createdPost.ID, retrievedPost.ID)
		assert.Equal(t, createdPost.Title, retrievedPost.Title)

		// Update post
		updateInput := &model.PostUpdateInput{
			Title:   stringPtr("Updated Post Title"),
			Content: stringPtr("Updated content"),
		}

		updatedPost, err := postService.UpdatePost(ctx, createdPost.ID, updateInput)
		require.NoError(t, err)
		assert.Equal(t, *updateInput.Title, updatedPost.Title)
		assert.Equal(t, *updateInput.Content, updatedPost.Content)
		assert.True(t, updatedPost.UpdatedAt.After(updatedPost.CreatedAt))

		// List posts
		filter := &model.PostFilter{
			AuthorID: &authorID,
		}
		pagination := &model.PaginationInput{
			First: intPtr(10),
		}

		postConnection, err := postService.ListPosts(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Len(t, postConnection.Edges, 1)
		assert.Equal(t, updatedPost.ID, postConnection.Edges[0].Node.ID)

		// Delete post
		err = postService.DeletePost(ctx, createdPost.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = postService.GetPost(ctx, createdPost.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, model.ErrPostNotFound)
	})

	t.Run("full comment lifecycle", func(t *testing.T) {
		// Create post first
		authorID := uuid.New()
		postInput := &model.PostInput{
			Title:           "Post with Comments",
			Content:         "This post will have comments",
			AuthorID:        authorID,
			CommentsEnabled: true,
		}

		post, err := postService.CreatePost(ctx, postInput)
		require.NoError(t, err)

		// Create root comment
		commenterID := uuid.New()
		commentInput := &model.CommentInput{
			PostID:   post.ID,
			Content:  "This is a root comment",
			AuthorID: commenterID,
		}

		rootComment, err := commentService.CreateComment(ctx, commentInput)
		require.NoError(t, err)
		require.NotNil(t, rootComment)
		assert.Equal(t, commentInput.Content, rootComment.Content)
		assert.Equal(t, commentInput.AuthorID, rootComment.AuthorID)
		assert.Equal(t, commentInput.PostID, rootComment.PostID)
		assert.Nil(t, rootComment.ParentID)
		assert.Equal(t, 0, rootComment.Depth)

		// Create child comment
		childCommentInput := &model.CommentInput{
			PostID:   post.ID,
			ParentID: &rootComment.ID,
			Content:  "This is a child comment",
			AuthorID: commenterID,
		}

		childComment, err := commentService.CreateComment(ctx, childCommentInput)
		require.NoError(t, err)
		assert.Equal(t, childCommentInput.Content, childComment.Content)
		assert.Equal(t, &rootComment.ID, childComment.ParentID)
		assert.Equal(t, 1, childComment.Depth)

		// Get comment
		retrievedComment, err := commentService.GetComment(ctx, rootComment.ID)
		require.NoError(t, err)
		assert.Equal(t, rootComment.ID, retrievedComment.ID)

		// List comments
		commentFilter := &model.CommentFilter{
			PostID: &post.ID,
		}
		commentPagination := &model.PaginationInput{
			First: intPtr(10),
		}

		commentConnection, err := commentService.ListComments(ctx, commentFilter, commentPagination)
		require.NoError(t, err)
		assert.Len(t, commentConnection.Edges, 2)

		// Get comments tree
		commentsTree, err := commentService.GetCommentsTree(ctx, post.ID, nil)
		require.NoError(t, err)
		assert.Len(t, commentsTree, 2)

		// Update comment
		updateCommentInput := &model.CommentUpdateInput{
			Content: stringPtr("Updated comment content"),
		}

		updatedComment, err := commentService.UpdateComment(ctx, rootComment.ID, updateCommentInput)
		require.NoError(t, err)
		assert.Equal(t, *updateCommentInput.Content, updatedComment.Content)

		// Delete child comment first
		err = commentService.DeleteComment(ctx, childComment.ID)
		require.NoError(t, err)

		// Delete root comment
		err = commentService.DeleteComment(ctx, rootComment.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = commentService.GetComment(ctx, rootComment.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, model.ErrCommentNotFound)
	})

	t.Run("validation errors", func(t *testing.T) {
		// Invalid post input
		invalidPostInput := &model.PostInput{
			Title:    "", // Empty title should fail
			Content:  "Valid content",
			AuthorID: uuid.New(),
		}

		_, err := postService.CreatePost(ctx, invalidPostInput)
		assert.Error(t, err)

		// Create valid post for comment tests
		validPost, err := postService.CreatePost(ctx, &model.PostInput{
			Title:           "Valid Post",
			Content:         "Valid content",
			AuthorID:        uuid.New(),
			CommentsEnabled: true,
		})
		require.NoError(t, err)

		// Invalid comment input
		invalidCommentInput := &model.CommentInput{
			PostID:   validPost.ID,
			Content:  "", // Empty content should fail
			AuthorID: uuid.New(),
		}

		_, err = commentService.CreateComment(ctx, invalidCommentInput)
		assert.Error(t, err)

		// Comment on non-existent post
		nonExistentCommentInput := &model.CommentInput{
			PostID:   uuid.New(), // Non-existent post
			Content:  "Valid content",
			AuthorID: uuid.New(),
		}

		_, err = commentService.CreateComment(ctx, nonExistentCommentInput)
		assert.Error(t, err)
	})

	t.Run("pagination", func(t *testing.T) {
		authorID := uuid.New()

		// Create multiple posts
		var createdPosts []*model.Post
		for i := 0; i < 5; i++ {
			postInput := &model.PostInput{
				Title:           fmt.Sprintf("Post %d", i+1),
				Content:         fmt.Sprintf("Content for post %d", i+1),
				AuthorID:        authorID,
				CommentsEnabled: true,
			}

			post, err := postService.CreatePost(ctx, postInput)
			require.NoError(t, err)
			createdPosts = append(createdPosts, post)
		}

		// Test pagination
		pagination := &model.PaginationInput{
			First: intPtr(3),
		}

		connection, err := postService.ListPosts(ctx, nil, pagination)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(connection.Edges), 3)
		assert.NotNil(t, connection.PageInfo)

		if len(connection.Edges) == 3 {
			assert.True(t, connection.PageInfo.HasNextPage)
			assert.NotEmpty(t, connection.PageInfo.EndCursor)
		}

		// Clean up
		for _, post := range createdPosts {
			_ = postService.DeletePost(ctx, post.ID)
		}
	})

	_ = cfg // Use config variable to avoid unused variable warning
}

func TestCommentDepthValidation_Integration(t *testing.T) {
	// Setup
	logger := zaptest.NewLogger(t)
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository()

	repos := &repository.Repositories{
		Post:    postRepo,
		Comment: commentRepo,
	}

	serviceManager := service.NewManager(repos, logger)
	services := serviceManager.GetServices()
	defer serviceManager.Close()

	postService := services.Post
	commentService := services.Comment

	ctx := context.Background()

	// Create a post
	post, err := postService.CreatePost(ctx, &model.PostInput{
		Title:           "Depth Test Post",
		Content:         "Testing comment depth limits",
		AuthorID:        uuid.New(),
		CommentsEnabled: true,
	})
	require.NoError(t, err)

	authorID := uuid.New()

	// Create a chain of nested comments
	var parentID *uuid.UUID
	var lastComment *model.Comment

	// Create comments up to a reasonable depth
	for depth := 0; depth < 10; depth++ {
		commentInput := &model.CommentInput{
			PostID:   post.ID,
			ParentID: parentID,
			Content:  fmt.Sprintf("Comment at depth %d", depth),
			AuthorID: authorID,
		}

		comment, err := commentService.CreateComment(ctx, commentInput)
		require.NoError(t, err)
		assert.Equal(t, depth, comment.Depth)

		parentID = &comment.ID
		lastComment = comment
	}

	// Verify the last comment has the correct depth
	assert.Equal(t, 9, lastComment.Depth)

	// Get the comments tree and verify structure
	commentsTree, err := commentService.GetCommentsTree(ctx, post.ID, nil)
	require.NoError(t, err)
	assert.Len(t, commentsTree, 10)

	// Verify depths are correct
	for i, comment := range commentsTree {
		assert.Equal(t, i, comment.Depth)
	}
}

func TestServiceHealthCheck_Integration(t *testing.T) {
	// Setup
	logger := zaptest.NewLogger(t)
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository()

	repos := &repository.Repositories{
		Post:    postRepo,
		Comment: commentRepo,
	}

	serviceManager := service.NewManager(repos, logger)
	defer serviceManager.Close()

	ctx := context.Background()

	// Test health check
	err := serviceManager.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test metrics
	metrics := serviceManager.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "subscription")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
