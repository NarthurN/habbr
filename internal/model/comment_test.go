package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   CommentInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: CommentInput{
				PostID:   uuid.New(),
				Content:  "This is a valid comment",
				AuthorID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "valid input with parent",
			input: CommentInput{
				PostID:   uuid.New(),
				ParentID: &[]uuid.UUID{uuid.New()}[0],
				Content:  "This is a reply comment",
				AuthorID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "empty content",
			input: CommentInput{
				PostID:   uuid.New(),
				Content:  "",
				AuthorID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name: "whitespace only content",
			input: CommentInput{
				PostID:   uuid.New(),
				Content:  "   ",
				AuthorID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name: "content too long",
			input: CommentInput{
				PostID:   uuid.New(),
				Content:  string(make([]rune, MaxCommentLength+1)),
				AuthorID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "content cannot exceed 2000 characters",
		},
		{
			name: "nil post ID",
			input: CommentInput{
				PostID:   uuid.Nil,
				Content:  "Valid content",
				AuthorID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "post_id is required",
		},
		{
			name: "nil author ID",
			input: CommentInput{
				PostID:   uuid.New(),
				Content:  "Valid content",
				AuthorID: uuid.Nil,
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

func TestNewComment(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	parentID := uuid.New()

	tests := []struct {
		name     string
		input    CommentInput
		depth    int
		expected func(*Comment)
	}{
		{
			name: "root comment",
			input: CommentInput{
				PostID:   postID,
				Content:  "Root comment",
				AuthorID: authorID,
			},
			depth: 0,
			expected: func(c *Comment) {
				assert.NotEqual(t, uuid.Nil, c.ID)
				assert.Equal(t, postID, c.PostID)
				assert.Nil(t, c.ParentID)
				assert.Equal(t, "Root comment", c.Content)
				assert.Equal(t, authorID, c.AuthorID)
				assert.Equal(t, 0, c.Depth)
				assert.True(t, c.IsRootComment())
				assert.NotNil(t, c.Children)
				assert.Len(t, c.Children, 0)
			},
		},
		{
			name: "child comment",
			input: CommentInput{
				PostID:   postID,
				ParentID: &parentID,
				Content:  "Child comment",
				AuthorID: authorID,
			},
			depth: 1,
			expected: func(c *Comment) {
				assert.NotEqual(t, uuid.Nil, c.ID)
				assert.Equal(t, postID, c.PostID)
				assert.NotNil(t, c.ParentID)
				assert.Equal(t, parentID, *c.ParentID)
				assert.Equal(t, "Child comment", c.Content)
				assert.Equal(t, authorID, c.AuthorID)
				assert.Equal(t, 1, c.Depth)
				assert.False(t, c.IsRootComment())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment := NewComment(tt.input, tt.depth)
			tt.expected(comment)
		})
	}
}

func TestComment_Update(t *testing.T) {
	comment := NewComment(CommentInput{
		PostID:   uuid.New(),
		Content:  "Original content",
		AuthorID: uuid.New(),
	}, 0)

	originalCreatedAt := comment.CreatedAt
	originalUpdatedAt := comment.UpdatedAt

	// Обновляем комментарий
	newContent := "Updated content"
	updateInput := CommentUpdateInput{
		Content: &newContent,
	}

	comment.Update(updateInput)

	assert.Equal(t, "Updated content", comment.Content)
	assert.Equal(t, originalCreatedAt, comment.CreatedAt)      // CreatedAt не должно изменяться
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt)) // UpdatedAt должно обновиться
}

func TestComment_CanBeRepliedTo(t *testing.T) {
	comment := &Comment{
		Depth: 5,
	}

	// В текущей реализации всегда можно ответить на комментарий
	assert.True(t, comment.CanBeRepliedTo())
}

func TestComment_AddChild(t *testing.T) {
	parent := NewComment(CommentInput{
		PostID:   uuid.New(),
		Content:  "Parent comment",
		AuthorID: uuid.New(),
	}, 0)

	child := NewComment(CommentInput{
		PostID:   parent.PostID,
		ParentID: &parent.ID,
		Content:  "Child comment",
		AuthorID: uuid.New(),
	}, 1)

	parent.AddChild(child)

	assert.Len(t, parent.Children, 1)
	assert.Equal(t, child.ID, parent.Children[0].ID)
}

func TestBuildCommentsTree(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()

	// Создаем несколько комментариев
	root1 := NewComment(CommentInput{
		PostID:   postID,
		Content:  "Root comment 1",
		AuthorID: authorID,
	}, 0)

	root2 := NewComment(CommentInput{
		PostID:   postID,
		Content:  "Root comment 2",
		AuthorID: authorID,
	}, 0)

	child1 := NewComment(CommentInput{
		PostID:   postID,
		ParentID: &root1.ID,
		Content:  "Child of root 1",
		AuthorID: authorID,
	}, 1)

	grandchild1 := NewComment(CommentInput{
		PostID:   postID,
		ParentID: &child1.ID,
		Content:  "Grandchild of root 1",
		AuthorID: authorID,
	}, 2)

	// Плоский список комментариев
	flatComments := []*Comment{root1, root2, child1, grandchild1}

	// Строим дерево
	tree := BuildCommentsTree(flatComments)

	// Проверяем структуру дерева
	require.Len(t, tree, 2) // Два корневых комментария

	// Проверяем первый корневой комментарий
	assert.Equal(t, root1.ID, tree[0].ID)
	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, child1.ID, tree[0].Children[0].ID)
	require.Len(t, tree[0].Children[0].Children, 1)
	assert.Equal(t, grandchild1.ID, tree[0].Children[0].Children[0].ID)

	// Проверяем второй корневой комментарий
	assert.Equal(t, root2.ID, tree[1].ID)
	assert.Len(t, tree[1].Children, 0)
}

func TestFlattenCommentsTree(t *testing.T) {
	// Создаем дерево комментариев
	root := &Comment{
		ID:      uuid.New(),
		Content: "Root",
		Depth:   0,
	}

	child1 := &Comment{
		ID:      uuid.New(),
		Content: "Child 1",
		Depth:   1,
	}

	child2 := &Comment{
		ID:      uuid.New(),
		Content: "Child 2",
		Depth:   1,
	}

	grandchild := &Comment{
		ID:      uuid.New(),
		Content: "Grandchild",
		Depth:   2,
	}

	// Строим дерево
	child1.Children = []*Comment{grandchild}
	root.Children = []*Comment{child1, child2}
	tree := []*Comment{root}

	// Сплющиваем дерево
	flat := FlattenCommentsTree(tree)

	// Проверяем результат
	require.Len(t, flat, 4)
	assert.Equal(t, root.ID, flat[0].ID)
	assert.Equal(t, child1.ID, flat[1].ID)
	assert.Equal(t, grandchild.ID, flat[2].ID)
	assert.Equal(t, child2.ID, flat[3].ID)
}

func TestCommentUpdateInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   CommentUpdateInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid update",
			input: CommentUpdateInput{
				Content: stringPtr("Updated content"),
			},
			wantErr: false,
		},
		{
			name: "empty content",
			input: CommentUpdateInput{
				Content: stringPtr(""),
			},
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name: "content too long",
			input: CommentUpdateInput{
				Content: stringPtr(string(make([]rune, MaxCommentLength+1))),
			},
			wantErr: true,
			errMsg:  "content cannot exceed 2000 characters",
		},
		{
			name:    "nil content",
			input:   CommentUpdateInput{},
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
