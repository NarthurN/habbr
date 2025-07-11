package converter

import (
	"github.com/NarthurN/habbr/internal/model"
	repomodel "github.com/NarthurN/habbr/internal/repository/model"
)

// CommentToRepo конвертирует доменную модель комментария в модель репозитория
func CommentToRepo(comment *model.Comment) *repomodel.Comment {
	if comment == nil {
		return nil
	}

	return &repomodel.Comment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		AuthorID:  comment.AuthorID,
		Depth:     comment.Depth,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

// CommentFromRepo конвертирует модель репозитория в доменную модель комментария
func CommentFromRepo(comment *repomodel.Comment) *model.Comment {
	if comment == nil {
		return nil
	}

	return &model.Comment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		AuthorID:  comment.AuthorID,
		Depth:     comment.Depth,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		Children:  make([]*model.Comment, 0), // Дочерние комментарии будут добавлены отдельно
	}
}

// CommentsFromRepo конвертирует слайс моделей репозитория в слайс доменных моделей комментариев
func CommentsFromRepo(comments []*repomodel.Comment) []*model.Comment {
	if comments == nil {
		return nil
	}

	result := make([]*model.Comment, len(comments))
	for i, comment := range comments {
		result[i] = CommentFromRepo(comment)
	}

	return result
}

// CommentFilterToRepo конвертирует доменный фильтр комментариев в фильтр репозитория
func CommentFilterToRepo(filter model.CommentFilter, pagination model.PaginationInput) repomodel.CommentFilter {
	repoFilter := repomodel.CommentFilter{
		PostID:   filter.PostID,
		ParentID: filter.ParentID,
		AuthorID: filter.AuthorID,
		MaxDepth: filter.MaxDepth,
		OrderBy:  "created_at",
		OrderDir: "asc", // Комментарии обычно сортируются по возрастанию времени
		Limit:    50,    // значение по умолчанию
		Offset:   0,
	}

	// Применяем пагинацию
	if pagination.First != nil {
		repoFilter.Limit = *pagination.First
		if repoFilter.Limit > 200 {
			repoFilter.Limit = 200 // максимальный лимит для комментариев
		}
	}

	if pagination.Last != nil {
		repoFilter.Limit = *pagination.Last
		if repoFilter.Limit > 200 {
			repoFilter.Limit = 200
		}
	}

	return repoFilter
}

// CommentWithChildrenFromRepo конвертирует модель репозитория с количеством дочерних комментариев
func CommentWithChildrenFromRepo(comment *repomodel.CommentWithChildren) (*model.Comment, int) {
	if comment == nil {
		return nil, 0
	}

	domainComment := CommentFromRepo(&comment.Comment)
	return domainComment, comment.ChildrenCount
}
