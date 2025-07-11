package converter

import (
	"errors"
	"testing"
	"time"

	"github.com/NarthurN/habbr/internal/api/graphql/generated"
	"github.com/NarthurN/habbr/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCommentToGraphQL(t *testing.T) {
	postID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	parentID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
	authorID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174002")
	commentID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174003")
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *model.Comment
		expected *generated.Comment
	}{
		{
			name:     "nil comment",
			input:    nil,
			expected: nil,
		},
		{
			name: "root comment",
			input: &model.Comment{
				ID:        commentID,
				PostID:    postID,
				ParentID:  nil,
				Content:   "Test Comment",
				AuthorID:  authorID,
				Depth:     0,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			expected: &generated.Comment{
				ID:        commentID.String(),
				PostID:    postID.String(),
				ParentID:  nil,
				Content:   "Test Comment",
				AuthorID:  authorID.String(),
				Depth:     0,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
		},
		{
			name: "child comment",
			input: &model.Comment{
				ID:        commentID,
				PostID:    postID,
				ParentID:  &parentID,
				Content:   "Child Comment",
				AuthorID:  authorID,
				Depth:     1,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			expected: &generated.Comment{
				ID:        commentID.String(),
				PostID:    postID.String(),
				ParentID:  stringPtr(parentID.String()),
				Content:   "Child Comment",
				AuthorID:  authorID.String(),
				Depth:     1,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CommentToGraphQL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommentInputFromGraphQL(t *testing.T) {
	postID := "123e4567-e89b-12d3-a456-426614174000"
	parentID := "123e4567-e89b-12d3-a456-426614174001"
	authorID := "123e4567-e89b-12d3-a456-426614174002"

	tests := []struct {
		name        string
		input       generated.CommentInput
		expected    *model.CommentInput
		expectError bool
	}{
		{
			name: "valid root comment",
			input: generated.CommentInput{
				PostID:   postID,
				ParentID: nil,
				Content:  "Test Comment",
				AuthorID: authorID,
			},
			expected: &model.CommentInput{
				PostID:   uuid.MustParse(postID),
				ParentID: nil,
				Content:  "Test Comment",
				AuthorID: uuid.MustParse(authorID),
			},
			expectError: false,
		},
		{
			name: "valid child comment",
			input: generated.CommentInput{
				PostID:   postID,
				ParentID: &parentID,
				Content:  "Child Comment",
				AuthorID: authorID,
			},
			expected: &model.CommentInput{
				PostID:   uuid.MustParse(postID),
				ParentID: &uuid.UUID{},
				Content:  "Child Comment",
				AuthorID: uuid.MustParse(authorID),
			},
			expectError: false,
		},
		{
			name: "invalid post ID",
			input: generated.CommentInput{
				PostID:   "invalid-uuid",
				Content:  "Test Comment",
				AuthorID: authorID,
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid author ID",
			input: generated.CommentInput{
				PostID:   postID,
				Content:  "Test Comment",
				AuthorID: "invalid-uuid",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid parent ID",
			input: generated.CommentInput{
				PostID:   postID,
				ParentID: stringPtr("invalid-uuid"),
				Content:  "Test Comment",
				AuthorID: authorID,
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CommentInputFromGraphQL(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expected != nil {
					assert.Equal(t, tt.expected.PostID, result.PostID)
					assert.Equal(t, tt.expected.Content, result.Content)
					assert.Equal(t, tt.expected.AuthorID, result.AuthorID)

					if tt.input.ParentID != nil {
						expectedParentID := uuid.MustParse(*tt.input.ParentID)
						assert.Equal(t, &expectedParentID, result.ParentID)
					} else {
						assert.Nil(t, result.ParentID)
					}
				}
			}
		})
	}
}

func TestCommentUpdateInputFromGraphQL(t *testing.T) {
	content := "Updated Content"

	tests := []struct {
		name     string
		input    generated.CommentUpdateInput
		expected *model.CommentUpdateInput
	}{
		{
			name:  "empty input",
			input: generated.CommentUpdateInput{},
			expected: &model.CommentUpdateInput{
				Content: nil,
			},
		},
		{
			name: "with content",
			input: generated.CommentUpdateInput{
				Content: &content,
			},
			expected: &model.CommentUpdateInput{
				Content: &content,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CommentUpdateInputFromGraphQL(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommentFilterFromGraphQL(t *testing.T) {
	authorID := "123e4567-e89b-12d3-a456-426614174000"
	maxDepth := 5

	tests := []struct {
		name        string
		input       *generated.CommentFilter
		expected    *model.CommentFilter
		expectError bool
	}{
		{
			name:     "nil filter",
			input:    nil,
			expected: &model.CommentFilter{},
		},
		{
			name:     "empty filter",
			input:    &generated.CommentFilter{},
			expected: &model.CommentFilter{},
		},
		{
			name: "filter with valid author ID",
			input: &generated.CommentFilter{
				AuthorID: &authorID,
				MaxDepth: &maxDepth,
			},
			expected: &model.CommentFilter{
				AuthorID: &uuid.UUID{},
				MaxDepth: &maxDepth,
			},
		},
		{
			name: "filter with invalid author ID",
			input: &generated.CommentFilter{
				AuthorID: stringPtr("invalid-uuid"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CommentFilterFromGraphQL(tt.input)

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
				assert.Equal(t, tt.expected.MaxDepth, result.MaxDepth)
			}
		})
	}
}

func TestCommentConnectionToGraphQL(t *testing.T) {
	comment1 := &model.Comment{
		ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		PostID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		Content:  "Comment 1",
		AuthorID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
		Depth:    0,
	}

	comment2 := &model.Comment{
		ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174003"),
		PostID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		Content:  "Comment 2",
		AuthorID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174004"),
		Depth:    1,
	}

	tests := []struct {
		name     string
		input    *model.CommentConnection
		expected *generated.CommentConnection
	}{
		{
			name:  "nil connection",
			input: nil,
			expected: &generated.CommentConnection{
				Edges:      []*generated.CommentEdge{},
				PageInfo:   &generated.PageInfo{},
				TotalCount: 0,
			},
		},
		{
			name: "connection with comments",
			input: &model.CommentConnection{
				Edges: []*model.CommentEdge{
					{Node: comment1, Cursor: "cursor1"},
					{Node: comment2, Cursor: "cursor2"},
				},
				PageInfo: &model.PageInfo{
					HasNextPage:     false,
					HasPreviousPage: true,
					StartCursor:     stringPtr("cursor1"),
					EndCursor:       stringPtr("cursor2"),
				},
			},
			expected: &generated.CommentConnection{
				Edges: []*generated.CommentEdge{
					{Node: CommentToGraphQL(comment1), Cursor: "cursor1"},
					{Node: CommentToGraphQL(comment2), Cursor: "cursor2"},
				},
				PageInfo: &generated.PageInfo{
					HasNextPage:     false,
					HasPreviousPage: true,
					StartCursor:     stringPtr("cursor1"),
					EndCursor:       stringPtr("cursor2"),
				},
				TotalCount: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CommentConnectionToGraphQL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommentResultToGraphQL(t *testing.T) {
	comment := &model.Comment{
		ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		PostID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		Content:  "Test Comment",
		AuthorID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
		Depth:    0,
	}

	tests := []struct {
		name     string
		comment  *model.Comment
		err      error
		expected *generated.CommentResult
	}{
		{
			name:    "success result",
			comment: comment,
			err:     nil,
			expected: &generated.CommentResult{
				Success: true,
				Comment: CommentToGraphQL(comment),
				Error:   nil,
			},
		},
		{
			name:    "error result",
			comment: nil,
			err:     errors.New("test error"),
			expected: &generated.CommentResult{
				Success: false,
				Comment: nil,
				Error:   stringPtr("test error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CommentResultToGraphQL(tt.comment, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBatchDeleteResultToGraphQL(t *testing.T) {
	id1 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	id2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	tests := []struct {
		name       string
		deletedIDs []uuid.UUID
		errors     []error
		expected   *generated.BatchDeleteResult
	}{
		{
			name:       "success with no errors",
			deletedIDs: []uuid.UUID{id1, id2},
			errors:     []error{},
			expected: &generated.BatchDeleteResult{
				Success:      true,
				DeletedCount: 2,
				DeletedIDs:   []string{id1.String(), id2.String()},
				Errors:       []string{},
			},
		},
		{
			name:       "partial success with errors",
			deletedIDs: []uuid.UUID{id1},
			errors:     []error{err1, err2},
			expected: &generated.BatchDeleteResult{
				Success:      false,
				DeletedCount: 1,
				DeletedIDs:   []string{id1.String()},
				Errors:       []string{"error 1", "error 2"},
			},
		},
		{
			name:       "complete failure",
			deletedIDs: []uuid.UUID{},
			errors:     []error{err1},
			expected: &generated.BatchDeleteResult{
				Success:      false,
				DeletedCount: 0,
				DeletedIDs:   []string{},
				Errors:       []string{"error 1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BatchDeleteResultToGraphQL(tt.deletedIDs, tt.errors)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommentEventToGraphQL(t *testing.T) {
	comment := &model.Comment{
		ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		PostID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		Content:  "Test Comment",
		AuthorID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
		Depth:    0,
	}

	tests := []struct {
		name        string
		eventType   string
		comment     *model.Comment
		expected    *generated.CommentEvent
		expectError bool
	}{
		{
			name:      "created event",
			eventType: "CREATED",
			comment:   comment,
			expected: &generated.CommentEvent{
				Type:    generated.CommentEventTypeCreated,
				Comment: CommentToGraphQL(comment),
				PostID:  comment.PostID.String(),
			},
			expectError: false,
		},
		{
			name:      "updated event",
			eventType: "UPDATED",
			comment:   comment,
			expected: &generated.CommentEvent{
				Type:    generated.CommentEventTypeUpdated,
				Comment: CommentToGraphQL(comment),
				PostID:  comment.PostID.String(),
			},
			expectError: false,
		},
		{
			name:      "deleted event",
			eventType: "DELETED",
			comment:   comment,
			expected: &generated.CommentEvent{
				Type:    generated.CommentEventTypeDeleted,
				Comment: CommentToGraphQL(comment),
				PostID:  comment.PostID.String(),
			},
			expectError: false,
		},
		{
			name:        "unknown event type",
			eventType:   "UNKNOWN",
			comment:     comment,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CommentEventToGraphQL(tt.eventType, tt.comment)

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
