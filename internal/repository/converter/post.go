package converter

import (
	"github.com/NarthurN/habbr/internal/model"
	repomodel "github.com/NarthurN/habbr/internal/repository/model"
)

// PostToRepo конвертирует доменную модель поста в модель репозитория
func PostToRepo(post *model.Post) *repomodel.Post {
	if post == nil {
		return nil
	}

	return &repomodel.Post{
		ID:              post.ID,
		Title:           post.Title,
		Content:         post.Content,
		AuthorID:        post.AuthorID,
		CommentsEnabled: post.CommentsEnabled,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
	}
}

// PostFromRepo конвертирует модель репозитория в доменную модель поста
func PostFromRepo(post *repomodel.Post) *model.Post {
	if post == nil {
		return nil
	}

	return &model.Post{
		ID:              post.ID,
		Title:           post.Title,
		Content:         post.Content,
		AuthorID:        post.AuthorID,
		CommentsEnabled: post.CommentsEnabled,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
	}
}

// PostsFromRepo конвертирует слайс моделей репозитория в слайс доменных моделей постов
func PostsFromRepo(posts []*repomodel.Post) []*model.Post {
	if posts == nil {
		return nil
	}

	result := make([]*model.Post, len(posts))
	for i, post := range posts {
		result[i] = PostFromRepo(post)
	}

	return result
}

// PostFilterToRepo конвертирует доменный фильтр постов в фильтр репозитория
func PostFilterToRepo(filter model.PostFilter, pagination model.PaginationInput) repomodel.PostFilter {
	repoFilter := repomodel.PostFilter{
		AuthorID:     filter.AuthorID,
		WithComments: filter.WithComments,
		OrderBy:      "created_at",
		OrderDir:     "desc",
		Limit:        20, // значение по умолчанию
		Offset:       0,
	}

	// Применяем пагинацию
	if pagination.First != nil {
		repoFilter.Limit = *pagination.First
		if repoFilter.Limit > 100 {
			repoFilter.Limit = 100 // максимальный лимит
		}
	}

	if pagination.Last != nil {
		repoFilter.Limit = *pagination.Last
		if repoFilter.Limit > 100 {
			repoFilter.Limit = 100
		}
	}

	// Для cursor-based пагинации offset будет вычислен отдельно

	return repoFilter
}

// PostWithCommentCountFromRepo конвертирует модель репозитория с количеством комментариев
func PostWithCommentCountFromRepo(post *repomodel.PostWithCommentCount) (*model.Post, int) {
	if post == nil {
		return nil, 0
	}

	domainPost := PostFromRepo(&post.Post)
	return domainPost, post.CommentCount
}
