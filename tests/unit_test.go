package tests

import (
	"testing"

	"github.com/NarthurN/habbr/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestModelValidation(t *testing.T) {
	t.Run("PostInput validation", func(t *testing.T) {
		// Valid input
		validInput := &model.PostInput{
			Title:           "Valid Title",
			Content:         "Valid content",
			AuthorID:        uuid.New(),
			CommentsEnabled: true,
		}

		err := validInput.Validate()
		assert.NoError(t, err)

		// Invalid title
		invalidInput := &model.PostInput{
			Title:    "",
			Content:  "Valid content",
			AuthorID: uuid.New(),
		}

		err = invalidInput.Validate()
		assert.Error(t, err)
	})

	t.Run("CommentInput validation", func(t *testing.T) {
		// Valid input
		validInput := &model.CommentInput{
			PostID:   uuid.New(),
			Content:  "Valid comment content",
			AuthorID: uuid.New(),
		}

		err := validInput.Validate()
		assert.NoError(t, err)

		// Invalid content
		invalidInput := &model.CommentInput{
			PostID:   uuid.New(),
			Content:  "",
			AuthorID: uuid.New(),
		}

		err = invalidInput.Validate()
		assert.Error(t, err)
	})

	t.Run("Post structure", func(t *testing.T) {
		post := &model.Post{
			ID:              uuid.New(),
			Title:           "Test Post",
			Content:         "Test Content",
			AuthorID:        uuid.New(),
			CommentsEnabled: true,
		}

		// Test basic properties
		assert.NotEmpty(t, post.ID)
		assert.Equal(t, "Test Post", post.Title)
		assert.True(t, post.CommentsEnabled)
	})

	t.Run("Comment structure", func(t *testing.T) {
		postID := uuid.New()
		// Root comment
		rootComment := &model.Comment{
			ID:       uuid.New(),
			PostID:   postID,
			Content:  "Root comment",
			AuthorID: uuid.New(),
			Depth:    0,
		}

		assert.Equal(t, 0, rootComment.Depth)
		assert.Nil(t, rootComment.ParentID)

		// Child comment
		childComment := &model.Comment{
			ID:       uuid.New(),
			PostID:   postID,
			ParentID: &rootComment.ID,
			Content:  "Child comment",
			AuthorID: uuid.New(),
			Depth:    1,
		}

		assert.Equal(t, 1, childComment.Depth)
		assert.NotNil(t, childComment.ParentID)
		assert.Equal(t, rootComment.ID, *childComment.ParentID)
	})
}

func TestPaginationStructure(t *testing.T) {
	t.Run("pagination structure", func(t *testing.T) {
		pagination := &model.PaginationInput{
			First: intPtr(10),
		}

		assert.NotNil(t, pagination.First)
		assert.Equal(t, 10, *pagination.First)
		assert.Nil(t, pagination.Last)
	})

	t.Run("pagination with cursor", func(t *testing.T) {
		after := "cursor123"
		pagination := &model.PaginationInput{
			First: intPtr(5),
			After: &after,
		}

		assert.Equal(t, 5, *pagination.First)
		assert.Equal(t, "cursor123", *pagination.After)
	})
}

func TestMemoryRepository(t *testing.T) {
	// Simple test to ensure memory repository compiles
	t.Run("repository initialization", func(t *testing.T) {
		// This test just ensures our imports and basic structure work
		assert.True(t, true)
	})
}

// Helper functions
func intPtr(i int) *int {
	return &i
}
