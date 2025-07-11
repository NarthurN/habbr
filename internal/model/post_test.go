package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   PostInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: PostInput{
				Title:           "Test Post",
				Content:         "This is a test post content",
				AuthorID:        uuid.New(),
				CommentsEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "empty title",
			input: PostInput{
				Title:           "",
				Content:         "Content",
				AuthorID:        uuid.New(),
				CommentsEnabled: true,
			},
			wantErr: true,
			errMsg:  "title cannot be empty",
		},
		{
			name: "whitespace only title",
			input: PostInput{
				Title:           "   ",
				Content:         "Content",
				AuthorID:        uuid.New(),
				CommentsEnabled: true,
			},
			wantErr: true,
			errMsg:  "title cannot be empty",
		},
		{
			name: "title too long",
			input: PostInput{
				Title:           string(make([]rune, 201)), // 201 characters
				Content:         "Content",
				AuthorID:        uuid.New(),
				CommentsEnabled: true,
			},
			wantErr: true,
			errMsg:  "title cannot exceed 200 characters",
		},
		{
			name: "empty content",
			input: PostInput{
				Title:           "Title",
				Content:         "",
				AuthorID:        uuid.New(),
				CommentsEnabled: true,
			},
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name: "content too long",
			input: PostInput{
				Title:           "Title",
				Content:         string(make([]rune, 50001)), // 50001 characters
				AuthorID:        uuid.New(),
				CommentsEnabled: true,
			},
			wantErr: true,
			errMsg:  "content cannot exceed 50000 characters",
		},
		{
			name: "nil author ID",
			input: PostInput{
				Title:           "Title",
				Content:         "Content",
				AuthorID:        uuid.Nil,
				CommentsEnabled: true,
			},
			wantErr: true,
			errMsg:  "author_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewPost(t *testing.T) {
	authorID := uuid.New()
	input := PostInput{
		Title:           "Test Post",
		Content:         "This is a test post",
		AuthorID:        authorID,
		CommentsEnabled: true,
	}

	post := NewPost(input)

	assert.NotEqual(t, uuid.Nil, post.ID)
	assert.Equal(t, "Test Post", post.Title)
	assert.Equal(t, "This is a test post", post.Content)
	assert.Equal(t, authorID, post.AuthorID)
	assert.True(t, post.CommentsEnabled)
	assert.False(t, post.CreatedAt.IsZero())
	assert.False(t, post.UpdatedAt.IsZero())
	assert.Equal(t, post.CreatedAt, post.UpdatedAt)
}

func TestPost_Update(t *testing.T) {
	// Создаем пост
	post := NewPost(PostInput{
		Title:           "Original Title",
		Content:         "Original Content",
		AuthorID:        uuid.New(),
		CommentsEnabled: true,
	})

	originalCreatedAt := post.CreatedAt
	originalUpdatedAt := post.UpdatedAt

	// Обновляем пост
	newTitle := "Updated Title"
	newContent := "Updated Content"
	commentsEnabled := false

	updateInput := PostUpdateInput{
		Title:           &newTitle,
		Content:         &newContent,
		CommentsEnabled: &commentsEnabled,
	}

	post.Update(updateInput)

	assert.Equal(t, "Updated Title", post.Title)
	assert.Equal(t, "Updated Content", post.Content)
	assert.False(t, post.CommentsEnabled)
	assert.Equal(t, originalCreatedAt, post.CreatedAt)      // CreatedAt не должно изменяться
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt)) // UpdatedAt должно обновиться
}

func TestPost_CanAddComments(t *testing.T) {
	tests := []struct {
		name            string
		commentsEnabled bool
		expected        bool
	}{
		{
			name:            "comments enabled",
			commentsEnabled: true,
			expected:        true,
		},
		{
			name:            "comments disabled",
			commentsEnabled: false,
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := &Post{
				CommentsEnabled: tt.commentsEnabled,
			}

			result := post.CanAddComments()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPostUpdateInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   PostUpdateInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid partial update",
			input: PostUpdateInput{
				Title: stringPtr("New Title"),
			},
			wantErr: false,
		},
		{
			name: "empty string title",
			input: PostUpdateInput{
				Title: stringPtr(""),
			},
			wantErr: true,
			errMsg:  "title cannot be empty",
		},
		{
			name: "title too long",
			input: PostUpdateInput{
				Title: stringPtr(string(make([]rune, 201))),
			},
			wantErr: true,
			errMsg:  "title cannot exceed 200 characters",
		},
		{
			name: "empty content",
			input: PostUpdateInput{
				Content: stringPtr(""),
			},
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name: "content too long",
			input: PostUpdateInput{
				Content: stringPtr(string(make([]rune, 50001))),
			},
			wantErr: true,
			errMsg:  "content cannot exceed 50000 characters",
		},
		{
			name:    "nil values",
			input:   PostUpdateInput{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Вспомогательная функция для создания указателя на строку
func stringPtr(s string) *string {
	return &s
}
