package converter

import (
	"fmt"

	"github.com/NarthurN/habbr/internal/api/graphql/generated"
	"github.com/NarthurN/habbr/internal/model"
	"github.com/google/uuid"
)

// CommentToGraphQL конвертирует domain модель Comment в GraphQL
func CommentToGraphQL(comment *model.Comment) *generated.Comment {
	if comment == nil {
		return nil
	}

	var parentID *string
	if comment.ParentID != nil {
		id := comment.ParentID.String()
		parentID = &id
	}

	return &generated.Comment{
		ID:        comment.ID.String(),
		PostID:    comment.PostID.String(),
		ParentID:  parentID,
		Content:   comment.Content,
		AuthorID:  comment.AuthorID.String(),
		Depth:     comment.Depth,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

// CommentInputFromGraphQL конвертирует GraphQL CommentInput в domain модель
func CommentInputFromGraphQL(input generated.CommentInput) (*model.CommentInput, error) {
	postID, err := uuid.Parse(input.PostID)
	if err != nil {
		return nil, err
	}

	authorID, err := uuid.Parse(input.AuthorID)
	if err != nil {
		return nil, err
	}

	var parentID *uuid.UUID
	if input.ParentID != nil {
		id, err := uuid.Parse(*input.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &id
	}

	return &model.CommentInput{
		PostID:   postID,
		ParentID: parentID,
		Content:  input.Content,
		AuthorID: authorID,
	}, nil
}

// CommentUpdateInputFromGraphQL конвертирует GraphQL CommentUpdateInput в domain модель
func CommentUpdateInputFromGraphQL(input generated.CommentUpdateInput) (*model.CommentUpdateInput, error) {
	return &model.CommentUpdateInput{
		Content: input.Content,
	}, nil
}

// CommentFilterFromGraphQL конвертирует GraphQL CommentFilter в domain модель
func CommentFilterFromGraphQL(filter *generated.CommentFilter) (*model.CommentFilter, error) {
	if filter == nil {
		return &model.CommentFilter{}, nil
	}

	result := &model.CommentFilter{}

	if filter.AuthorID != nil {
		authorID, err := uuid.Parse(*filter.AuthorID)
		if err != nil {
			return nil, err
		}
		result.AuthorID = &authorID
	}

	if filter.MaxDepth != nil {
		result.MaxDepth = filter.MaxDepth
	}

	return result, nil
}

// CommentConnectionToGraphQL конвертирует domain CommentConnection в GraphQL
func CommentConnectionToGraphQL(conn *model.CommentConnection) *generated.CommentConnection {
	if conn == nil {
		return &generated.CommentConnection{
			Edges:      []*generated.CommentEdge{},
			PageInfo:   &generated.PageInfo{},
			TotalCount: 0,
		}
	}

	edges := make([]*generated.CommentEdge, len(conn.Edges))
	for i, edge := range conn.Edges {
		edges[i] = &generated.CommentEdge{
			Node:   CommentToGraphQL(edge.Node),
			Cursor: edge.Cursor,
		}
	}

	return &generated.CommentConnection{
		Edges: edges,
		PageInfo: &generated.PageInfo{
			HasNextPage:     conn.PageInfo.HasNextPage,
			HasPreviousPage: conn.PageInfo.HasPreviousPage,
			StartCursor:     conn.PageInfo.StartCursor,
			EndCursor:       conn.PageInfo.EndCursor,
		},
		TotalCount: len(conn.Edges),
	}
}

// CommentResultToGraphQL конвертирует результат операции с комментарием в GraphQL
func CommentResultToGraphQL(comment *model.Comment, err error) *generated.CommentResult {
	if err != nil {
		return &generated.CommentResult{
			Success: false,
			Comment: nil,
			Error:   stringPtr(err.Error()),
		}
	}

	return &generated.CommentResult{
		Success: true,
		Comment: CommentToGraphQL(comment),
		Error:   nil,
	}
}

// BatchDeleteResultToGraphQL конвертирует результат массового удаления в GraphQL
func BatchDeleteResultToGraphQL(deletedIDs []uuid.UUID, errors []error) *generated.BatchDeleteResult {
	success := len(errors) == 0
	deletedCount := len(deletedIDs)

	// Конвертируем ID в строки
	stringIDs := make([]string, len(deletedIDs))
	for i, id := range deletedIDs {
		stringIDs[i] = id.String()
	}

	// Конвертируем ошибки в строки
	errorMessages := make([]string, len(errors))
	for i, err := range errors {
		if err != nil {
			errorMessages[i] = err.Error()
		}
	}

	return &generated.BatchDeleteResult{
		Success:      success,
		DeletedCount: deletedCount,
		DeletedIDs:   stringIDs,
		Errors:       errorMessages,
	}
}

// CommentEventToGraphQL конвертирует событие комментария в GraphQL
func CommentEventToGraphQL(eventType string, comment *model.Comment) (*generated.CommentEvent, error) {
	var gqlEventType generated.CommentEventType

	switch eventType {
	case "CREATED":
		gqlEventType = generated.CommentEventTypeCreated
	case "UPDATED":
		gqlEventType = generated.CommentEventTypeUpdated
	case "DELETED":
		gqlEventType = generated.CommentEventTypeDeleted
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}

	return &generated.CommentEvent{
		Type:    gqlEventType,
		Comment: CommentToGraphQL(comment),
		PostID:  comment.PostID.String(),
	}, nil
}
