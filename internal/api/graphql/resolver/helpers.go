package resolver

import (
	"context"
	"time"

	"github.com/NarthurN/habbr/internal/api/graphql/converter"
	"github.com/NarthurN/habbr/internal/api/graphql/generated"
	"github.com/NarthurN/habbr/internal/service"
	"github.com/google/uuid"
)

// getCurrentPostStats получает текущую статистику поста
func getCurrentPostStats(ctx context.Context, services *service.Services, postID uuid.UUID) (*generated.PostStats, error) {
	// Получаем пост для проверки существования
	post, err := services.Post.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Получаем количество комментариев
	filter := generated.CommentFilter{}
	domainFilter, _ := converter.CommentFilterFromGraphQL(&filter)
	domainFilter.PostID = &postID

	pagination := converter.PaginationFromGraphQL(nil, nil, nil, nil)
	comments, err := services.Comment.ListComments(ctx, *domainFilter, *pagination)
	if err != nil {
		return nil, err
	}

	// Находим последний комментарий
	var lastCommentAt *time.Time
	if len(comments.Edges) > 0 {
		for _, edge := range comments.Edges {
			if lastCommentAt == nil || edge.Node.CreatedAt.After(*lastCommentAt) {
				lastCommentAt = &edge.Node.CreatedAt
			}
		}
	}

	return &generated.PostStats{
		TotalComments:   len(comments.Edges),
		CommentsEnabled: post.CommentsEnabled,
		LastCommentAt:   lastCommentAt,
	}, nil
}
