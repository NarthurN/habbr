package converter

import (
	"testing"
	"time"

	"github.com/NarthurN/habbr/internal/api/graphql/generated"
	"github.com/NarthurN/habbr/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPostToGraphQL(t *testing.T) {
	tests := []struct {
		name     string
		input    *model.Post
		expected *generated.Post
	}{
		{
			name:     "nil post",
			input:    nil,
			expected: nil,
		},
		{
			name: "valid post",
			input: &model.Post{
				ID:              uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Title:           "Test Post",
				Content:         "Test Content",
				AuthorID:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
				CommentsEnabled: true,
				CreatedAt:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:       time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			expected: &generated.Post{
				ID:              "123e4567-e89b-12d3-a456-426614174000",
				Title:           "Test Post",
				Content:         "Test Content",
				AuthorID:        "123e4567-e89b-12d3-a456-426614174001",
				CommentsEnabled: true,
				CreatedAt:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:       time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PostToGraphQL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPostInputFromGraphQL(t *testing.T) {
	tests := []struct {
		name        string
		input       generated.PostInput
		expected    *model.PostInput
		expectError bool
	}{
		{
			name: "valid input",
			input: generated.PostInput{
				Title:           "Test Post",
				Content:         "Test Content",
				AuthorID:        "123e4567-e89b-12d3-a456-426614174000",
				CommentsEnabled: true,
			},
			expected: &model.PostInput{
				Title:           "Test Post",
				Content:         "Test Content",
				AuthorID:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				CommentsEnabled: true,
			},
			expectError: false,
		},
		{
			name: "invalid author ID",
			input: generated.PostInput{
				Title:           "Test Post",
				Content:         "Test Content",
				AuthorID:        "invalid-uuid",
				CommentsEnabled: true,
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PostInputFromGraphQL(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPostUpdateInputFromGraphQL(t *testing.T) {
	title := "Updated Title"
	content := "Updated Content"
	commentsEnabled := false

	tests := []struct {
		name     string
		input    generated.PostUpdateInput
		expected *model.PostUpdateInput
	}{
		{
			name:  "empty input",
			input: generated.PostUpdateInput{},
			expected: &model.PostUpdateInput{
				Title:           nil,
				Content:         nil,
				CommentsEnabled: nil,
			},
		},
		{
			name: "full input",
			input: generated.PostUpdateInput{
				Title:           &title,
				Content:         &content,
				CommentsEnabled: &commentsEnabled,
			},
			expected: &model.PostUpdateInput{
				Title:           &title,
				Content:         &content,
				CommentsEnabled: &commentsEnabled,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PostUpdateInputFromGraphQL(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPostFilterFromGraphQL(t *testing.T) {
	authorID := "123e4567-e89b-12d3-a456-426614174000"
	title := "Test"
	content := "Content"
	commentsEnabled := true

	tests := []struct {
		name        string
		input       *generated.PostFilter
		expected    *model.PostFilter
		expectError bool
	}{
		{
			name:     "nil filter",
			input:    nil,
			expected: &model.PostFilter{},
		},
		{
			name:     "empty filter",
			input:    &generated.PostFilter{},
			expected: &model.PostFilter{},
		},
		{
			name: "filter with valid author ID",
			input: &generated.PostFilter{
				AuthorID:        &authorID,
				Title:           &title,
				Content:         &content,
				CommentsEnabled: &commentsEnabled,
			},
			expected: &model.PostFilter{
				AuthorID:     &uuid.UUID{},
				WithComments: &commentsEnabled,
			},
		},
		{
			name: "filter with invalid author ID",
			input: &generated.PostFilter{
				AuthorID: testStringPtr("invalid-uuid"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PostFilterFromGraphQL(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.input != nil && tt.input.AuthorID != nil && *tt.input.AuthorID == authorID {
					expectedUUID := uuid.MustParse(authorID)
					assert.Equal(t, &expectedUUID, result.AuthorID)
				} else {
					assert.Equal(t, tt.expected.AuthorID, result.AuthorID)
				}
				assert.Equal(t, tt.expected.WithComments, result.WithComments)
			}
		})
	}
}

func TestPostConnectionToGraphQL(t *testing.T) {
	post1 := &model.Post{
		ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Title:    "Post 1",
		Content:  "Content 1",
		AuthorID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
	}

	post2 := &model.Post{
		ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
		Title:    "Post 2",
		Content:  "Content 2",
		AuthorID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174003"),
	}

	tests := []struct {
		name     string
		input    *model.PostConnection
		expected *generated.PostConnection
	}{
		{
			name:  "nil connection",
			input: nil,
			expected: &generated.PostConnection{
				Edges:      []*generated.PostEdge{},
				PageInfo:   &generated.PageInfo{},
				TotalCount: 0,
			},
		},
		{
			name: "connection with posts",
			input: &model.PostConnection{
				Edges: []*model.PostEdge{
					{Node: post1, Cursor: "cursor1"},
					{Node: post2, Cursor: "cursor2"},
				},
				PageInfo: &model.PageInfo{
					HasNextPage:     true,
					HasPreviousPage: false,
					StartCursor:     testStringPtr("cursor1"),
					EndCursor:       testStringPtr("cursor2"),
				},
			},
			expected: &generated.PostConnection{
				Edges: []*generated.PostEdge{
					{Node: PostToGraphQL(post1), Cursor: "cursor1"},
					{Node: PostToGraphQL(post2), Cursor: "cursor2"},
				},
				PageInfo: &generated.PageInfo{
					HasNextPage:     true,
					HasPreviousPage: false,
					StartCursor:     testStringPtr("cursor1"),
					EndCursor:       testStringPtr("cursor2"),
				},
				TotalCount: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PostConnectionToGraphQL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPostResultToGraphQL(t *testing.T) {
	post := &model.Post{
		ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Title:    "Test Post",
		Content:  "Test Content",
		AuthorID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
	}

	tests := []struct {
		name     string
		post     *model.Post
		err      error
		expected *generated.PostResult
	}{
		{
			name: "success result",
			post: post,
			err:  nil,
			expected: &generated.PostResult{
				Success: true,
				Post:    PostToGraphQL(post),
				Error:   nil,
			},
		},
		{
			name: "error result",
			post: nil,
			err:  assert.AnError,
			expected: &generated.PostResult{
				Success: false,
				Post:    nil,
				Error:   testStringPtr(assert.AnError.Error()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PostResultToGraphQL(tt.post, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeleteResultToGraphQL(t *testing.T) {
	deletedID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name      string
		deletedID uuid.UUID
		err       error
		expected  *generated.DeleteResult
	}{
		{
			name:      "success result",
			deletedID: deletedID,
			err:       nil,
			expected: &generated.DeleteResult{
				Success:   true,
				DeletedID: testStringPtr(deletedID.String()),
				Error:     nil,
			},
		},
		{
			name:      "error result",
			deletedID: uuid.Nil,
			err:       assert.AnError,
			expected: &generated.DeleteResult{
				Success:   false,
				DeletedID: nil,
				Error:     testStringPtr(assert.AnError.Error()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeleteResultToGraphQL(tt.deletedID, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationFromGraphQL(t *testing.T) {
	first := 10
	last := 5
	after := "cursor1"
	before := "cursor2"

	tests := []struct {
		name     string
		first    *int
		last     *int
		after    *string
		before   *string
		expected *model.PaginationInput
	}{
		{
			name:   "nil values",
			first:  nil,
			last:   nil,
			after:  nil,
			before: nil,
			expected: &model.PaginationInput{
				First:  nil,
				Last:   nil,
				After:  nil,
				Before: nil,
			},
		},
		{
			name:   "all values",
			first:  &first,
			last:   &last,
			after:  &after,
			before: &before,
			expected: &model.PaginationInput{
				First:  &first,
				Last:   &last,
				After:  &after,
				Before: &before,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PaginationFromGraphQL(tt.first, tt.last, tt.after, tt.before)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseID(t *testing.T) {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name        string
		input       string
		expected    uuid.UUID
		expectError bool
	}{
		{
			name:        "valid UUID",
			input:       validUUID,
			expected:    uuid.MustParse(validUUID),
			expectError: false,
		},
		{
			name:        "invalid UUID",
			input:       "invalid-uuid",
			expected:    uuid.Nil,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    uuid.Nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseID(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// testStringPtr - вспомогательная функция для создания указателя на строку в тестах
func testStringPtr(s string) *string {
	return &s
}
