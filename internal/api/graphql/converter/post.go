package converter

import (
	"github.com/NarthurN/habbr/internal/api/graphql/generated"
	"github.com/NarthurN/habbr/internal/model"
	"github.com/google/uuid"
)

// PostToGraphQL конвертирует domain модель Post в GraphQL
func PostToGraphQL(post *model.Post) *generated.Post {
	if post == nil {
		return nil
	}

	return &generated.Post{
		ID:              post.ID.String(),
		Title:           post.Title,
		Content:         post.Content,
		AuthorID:        post.AuthorID.String(),
		CommentsEnabled: post.CommentsEnabled,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
	}
}

// PostInputFromGraphQL конвертирует GraphQL PostInput в domain модель
func PostInputFromGraphQL(input generated.PostInput) (*model.PostInput, error) {
	authorID, err := uuid.Parse(input.AuthorID)
	if err != nil {
		return nil, err
	}

	return &model.PostInput{
		Title:           input.Title,
		Content:         input.Content,
		AuthorID:        authorID,
		CommentsEnabled: input.CommentsEnabled,
	}, nil
}

// PostUpdateInputFromGraphQL конвертирует GraphQL PostUpdateInput в domain модель
func PostUpdateInputFromGraphQL(input generated.PostUpdateInput) (*model.PostUpdateInput, error) {
	return &model.PostUpdateInput{
		Title:           input.Title,
		Content:         input.Content,
		CommentsEnabled: input.CommentsEnabled,
	}, nil
}

// PostFilterFromGraphQL конвертирует GraphQL PostFilter в domain модель
func PostFilterFromGraphQL(filter *generated.PostFilter) (*model.PostFilter, error) {
	if filter == nil {
		return &model.PostFilter{}, nil
	}

	result := &model.PostFilter{}

	if filter.AuthorID != nil {
		authorID, err := uuid.Parse(*filter.AuthorID)
		if err != nil {
			return nil, err
		}
		result.AuthorID = &authorID
	}

	if filter.CommentsEnabled != nil {
		result.WithComments = filter.CommentsEnabled
	}

	return result, nil
}

// PostConnectionToGraphQL конвертирует domain PostConnection в GraphQL
func PostConnectionToGraphQL(conn *model.PostConnection) *generated.PostConnection {
	if conn == nil {
		return &generated.PostConnection{
			Edges:      []*generated.PostEdge{},
			PageInfo:   &generated.PageInfo{},
			TotalCount: 0,
		}
	}

	edges := make([]*generated.PostEdge, len(conn.Edges))
	for i, edge := range conn.Edges {
		edges[i] = &generated.PostEdge{
			Node:   PostToGraphQL(edge.Node),
			Cursor: edge.Cursor,
		}
	}

	return &generated.PostConnection{
		Edges: edges,
		PageInfo: &generated.PageInfo{
			HasNextPage:     conn.PageInfo.HasNextPage,
			HasPreviousPage: conn.PageInfo.HasPreviousPage,
			StartCursor:     conn.PageInfo.StartCursor,
			EndCursor:       conn.PageInfo.EndCursor,
		},
		TotalCount: len(conn.Edges), // используем количество элементов
	}
}

// PostResultToGraphQL конвертирует результат операции с постом в GraphQL
func PostResultToGraphQL(post *model.Post, err error) *generated.PostResult {
	if err != nil {
		return &generated.PostResult{
			Success: false,
			Post:    nil,
			Error:   stringPtr(err.Error()),
		}
	}

	return &generated.PostResult{
		Success: true,
		Post:    PostToGraphQL(post),
		Error:   nil,
	}
}

// DeleteResultToGraphQL конвертирует результат удаления в GraphQL
func DeleteResultToGraphQL(deletedID uuid.UUID, err error) *generated.DeleteResult {
	if err != nil {
		return &generated.DeleteResult{
			Success:   false,
			DeletedID: nil,
			Error:     stringPtr(err.Error()),
		}
	}

	id := deletedID.String()
	return &generated.DeleteResult{
		Success:   true,
		DeletedID: &id,
		Error:     nil,
	}
}

// PaginationFromGraphQL конвертирует GraphQL пагинацию в domain модель
func PaginationFromGraphQL(first, last *int, after, before *string) *model.PaginationInput {
	return &model.PaginationInput{
		First:  first,
		Last:   last,
		After:  after,
		Before: before,
	}
}

// stringPtr возвращает указатель на строку
func stringPtr(s string) *string {
	return &s
}

// ParseID парсит строковый ID в UUID
func ParseID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
