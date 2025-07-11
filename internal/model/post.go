package model

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Post представляет доменную модель поста в системе.
// Это основная сущность, которая содержит информацию о публикации пользователя.
//
// Пост может иметь комментарии, которые включаются или отключаются через поле CommentsEnabled.
// Все посты имеют уникальный идентификатор UUID и привязаны к автору через AuthorID.
//
// Пример использования:
//   post := NewPost(PostInput{
//       Title: "Заголовок поста",
//       Content: "Содержимое поста",
//       AuthorID: authorID,
//       CommentsEnabled: true,
//   })
type Post struct {
	// ID - уникальный идентификатор поста в формате UUID
	ID uuid.UUID `json:"id"`

	// Title - заголовок поста, максимум 200 символов
	Title string `json:"title"`

	// Content - основное содержимое поста, максимум 50000 символов
	Content string `json:"content"`

	// AuthorID - идентификатор автора поста в формате UUID
	AuthorID uuid.UUID `json:"author_id"`

	// CommentsEnabled - флаг, разрешены ли комментарии к этому посту
	CommentsEnabled bool `json:"comments_enabled"`

	// CreatedAt - время создания поста
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt - время последнего обновления поста
	UpdatedAt time.Time `json:"updated_at"`
}

// PostInput представляет входные данные для создания нового поста.
//
// Все поля являются обязательными, кроме CommentsEnabled (по умолчанию false).
// Валидация происходит через метод Validate().
//
// Пример использования:
//   input := PostInput{
//       Title: "Мой новый пост",
//       Content: "Интересное содержимое поста",
//       AuthorID: userID,
//       CommentsEnabled: true,
//   }
//   if err := input.Validate(); err != nil {
//       return err
//   }
type PostInput struct {
	// Title - заголовок поста, обязательное поле, от 1 до 200 символов
	Title string `json:"title"`

	// Content - содержимое поста, обязательное поле, от 1 до 50000 символов
	Content string `json:"content"`

	// AuthorID - идентификатор автора, обязательное поле
	AuthorID uuid.UUID `json:"author_id"`

	// CommentsEnabled - разрешены ли комментарии, по умолчанию false
	CommentsEnabled bool `json:"comments_enabled"`
}

// PostUpdateInput представляет входные данные для обновления существующего поста.
//
// Все поля являются опциональными (указатели). Обновляются только те поля,
// которые не равны nil. Валидация происходит через метод Validate().
//
// Пример использования:
//   newTitle := "Обновленный заголовок"
//   update := PostUpdateInput{
//       Title: &newTitle,
//       CommentsEnabled: &true,
//   }
type PostUpdateInput struct {
	// Title - новый заголовок поста, опциональное поле
	Title *string `json:"title,omitempty"`

	// Content - новое содержимое поста, опциональное поле
	Content *string `json:"content,omitempty"`

	// CommentsEnabled - новое значение разрешения комментариев, опциональное поле
	CommentsEnabled *bool `json:"comments_enabled,omitempty"`
}

// PostFilter представляет фильтры для поиска и выборки постов.
//
// Используется в операциях списочного получения постов для ограничения результатов.
// Все поля являются опциональными.
//
// Пример использования:
//   filter := PostFilter{
//       AuthorID: &userID,
//       WithComments: &true,
//   }
type PostFilter struct {
	// AuthorID - фильтр по автору поста, если указан, возвращаются только посты данного автора
	AuthorID *uuid.UUID `json:"author_id,omitempty"`

	// WithComments - фильтр по наличию комментариев:
	// true - только посты с включенными комментариями
	// false - только посты с отключенными комментариями
	// nil - все посты независимо от настройки комментариев
	WithComments *bool `json:"with_comments,omitempty"`
}

// PaginationInput представляет параметры пагинации для cursor-based подхода.
//
// Поддерживает как прямую (first/after), так и обратную (last/before) пагинацию.
// Cursor-based пагинация обеспечивает стабильные результаты даже при изменении данных.
//
// Примеры использования:
//   // Первые 10 записей
//   pagination := PaginationInput{First: &10}
//
//   // Следующие 10 записей после cursor
//   pagination := PaginationInput{First: &10, After: &cursor}
//
//   // Последние 5 записей перед cursor
//   pagination := PaginationInput{Last: &5, Before: &cursor}
type PaginationInput struct {
	// First - количество записей с начала (для прямой пагинации)
	First *int `json:"first,omitempty"`

	// After - cursor, после которого нужно получить записи (для прямой пагинации)
	After *string `json:"after,omitempty"`

	// Last - количество записей с конца (для обратной пагинации)
	Last *int `json:"last,omitempty"`

	// Before - cursor, перед которым нужно получить записи (для обратной пагинации)
	Before *string `json:"before,omitempty"`
}

// PostConnection представляет результат пагинированного запроса постов.
//
// Следует стандарту GraphQL Cursor Connections Specification.
// Содержит сами данные (edges) и метаинформацию о пагинации (pageInfo).
//
// Пример использования:
//   connection, err := postService.ListPosts(ctx, filter, pagination)
//   for _, edge := range connection.Edges {
//       fmt.Printf("Post: %s, Cursor: %s\n", edge.Node.Title, edge.Cursor)
//   }
//   if connection.PageInfo.HasNextPage {
//       // Есть еще страницы для загрузки
//   }
type PostConnection struct {
	// Edges - массив ребер, каждое содержит пост и его cursor
	Edges []*PostEdge `json:"edges"`

	// PageInfo - информация о пагинации и доступности соседних страниц
	PageInfo *PageInfo `json:"page_info"`
}

// PostEdge представляет одно ребро в connection - комбинацию данных и cursor.
//
// Cursor используется для получения следующих или предыдущих страниц.
// Это непрозрачная строка, которую не следует парсить на клиенте.
type PostEdge struct {
	// Node - сам объект поста
	Node *Post `json:"node"`

	// Cursor - уникальный идентификатор позиции этого поста в результатах
	Cursor string `json:"cursor"`
}

// PageInfo содержит метаинформацию о состоянии пагинации.
//
// Используется клиентом для определения возможности навигации вперед/назад
// и получения cursors для следующих запросов.
type PageInfo struct {
	// HasNextPage - есть ли следующая страница (можно ли получить больше записей вперед)
	HasNextPage bool `json:"has_next_page"`

	// HasPreviousPage - есть ли предыдущая страница (можно ли получить записи назад)
	HasPreviousPage bool `json:"has_previous_page"`

	// StartCursor - cursor первого элемента в текущей странице (может быть nil)
	StartCursor *string `json:"start_cursor"`

	// EndCursor - cursor последнего элемента в текущей странице (может быть nil)
	EndCursor *string `json:"end_cursor"`
}

// Validate проверяет валидность входных данных для создания поста.
//
// Выполняет следующие проверки:
// - Заголовок не пустой и не превышает 200 символов
// - Содержимое не пустое и не превышает 50000 символов
// - AuthorID не является пустым UUID
//
// Возвращает:
//   - nil если все данные валидны
//   - error с описанием первой найденной ошибки валидации
//
// Пример использования:
//   input := PostInput{Title: "Test", Content: "Content", AuthorID: uuid.New()}
//   if err := input.Validate(); err != nil {
//       log.Printf("Ошибка валидации: %v", err)
//       return err
//   }
func (p *PostInput) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("title cannot be empty")
	}

	if len(p.Title) > 200 {
		return errors.New("title cannot exceed 200 characters")
	}

	if strings.TrimSpace(p.Content) == "" {
		return errors.New("content cannot be empty")
	}

	if len(p.Content) > 50000 {
		return errors.New("content cannot exceed 50000 characters")
	}

	if p.AuthorID == uuid.Nil {
		return errors.New("author_id is required")
	}

	return nil
}

// Validate проверяет валидность данных для обновления поста.
//
// Проверяет только те поля, которые не равны nil:
// - Если Title указан - проверяет что он не пустой и не превышает 200 символов
// - Если Content указан - проверяет что он не пустой и не превышает 50000 символов
// - CommentsEnabled не валидируется (любое булево значение допустимо)
//
// Возвращает:
//   - nil если все указанные данные валидны
//   - error с описанием первой найденной ошибки валидации
//
// Пример использования:
//   newTitle := "Новый заголовок"
//   update := PostUpdateInput{Title: &newTitle}
//   if err := update.Validate(); err != nil {
//       return fmt.Errorf("ошибка валидации обновления: %w", err)
//   }
func (p *PostUpdateInput) Validate() error {
	if p.Title != nil {
		if strings.TrimSpace(*p.Title) == "" {
			return errors.New("title cannot be empty")
		}

		if len(*p.Title) > 200 {
			return errors.New("title cannot exceed 200 characters")
		}
	}

	if p.Content != nil {
		if strings.TrimSpace(*p.Content) == "" {
			return errors.New("content cannot be empty")
		}

		if len(*p.Content) > 50000 {
			return errors.New("content cannot exceed 50000 characters")
		}
	}

	return nil
}

// NewPost создает новый пост из входных данных с автоматической генерацией ID и временных меток.
//
// Функция выполняет следующие действия:
// - Генерирует новый UUID для поста
// - Обрезает пробелы в начале и конце заголовка и содержимого
// - Устанавливает текущее время как CreatedAt и UpdatedAt
// - Копирует остальные поля из входных данных
//
// Параметры:
//   - input: валидированные входные данные для создания поста
//
// Возвращает:
//   - *Post: новый пост, готовый для сохранения в репозитории
//
// Примечание: Функция НЕ выполняет валидацию входных данных.
// Валидацию следует выполнить заранее через input.Validate().
//
// Пример использования:
//   input := PostInput{
//       Title: "  Заголовок с пробелами  ",
//       Content: "Содержимое поста",
//       AuthorID: uuid.New(),
//       CommentsEnabled: true,
//   }
//   if err := input.Validate(); err != nil {
//       return nil, err
//   }
//   post := NewPost(input)
//   // post.Title теперь "Заголовок с пробелами" (без пробелов по краям)
func NewPost(input PostInput) *Post {
	now := time.Now()
	return &Post{
		ID:              uuid.New(),
		Title:           strings.TrimSpace(input.Title),
		Content:         strings.TrimSpace(input.Content),
		AuthorID:        input.AuthorID,
		CommentsEnabled: input.CommentsEnabled,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Update обновляет существующий пост новыми данными.
//
// Метод применяет изменения только к тем полям, которые указаны в input (не равны nil).
// Автоматически обновляет временную метку UpdatedAt.
// Обрезает пробелы в начале и конце текстовых полей.
//
// Параметры:
//   - input: данные для обновления, поля равные nil игнорируются
//
// Побочные эффекты:
//   - Изменяет поля поста в соответствии с переданными данными
//   - Обновляет поле UpdatedAt на текущее время
//
// Примечание: Метод НЕ выполняет валидацию входных данных.
// Валидацию следует выполнить заранее через input.Validate().
//
// Пример использования:
//   newTitle := "Обновленный заголовок"
//   enabled := false
//   update := PostUpdateInput{
//       Title: &newTitle,
//       CommentsEnabled: &enabled,
//       // Content не указан, поэтому останется прежним
//   }
//   post.Update(update)
//   // post.Title = "Обновленный заголовок"
//   // post.CommentsEnabled = false
//   // post.Content остался прежним
//   // post.UpdatedAt = текущее время
func (p *Post) Update(input PostUpdateInput) {
	if input.Title != nil {
		p.Title = strings.TrimSpace(*input.Title)
	}

	if input.Content != nil {
		p.Content = strings.TrimSpace(*input.Content)
	}

	if input.CommentsEnabled != nil {
		p.CommentsEnabled = *input.CommentsEnabled
	}

	p.UpdatedAt = time.Now()
}

// CanAddComments проверяет, разрешено ли добавление комментариев к этому посту.
//
// Метод инкапсулирует бизнес-логику проверки возможности комментирования.
// В текущей реализации проверяется только флаг CommentsEnabled,
// но в будущем здесь могут быть дополнительные проверки (например, архивирования поста).
//
// Возвращает:
//   - true если комментарии разрешены
//   - false если комментарии запрещены
//
// Пример использования:
//   if post.CanAddComments() {
//       // Можно создавать комментарий к посту
//       comment := NewComment(commentInput)
//   } else {
//       return errors.New("комментарии к этому посту отключены")
//   }
func (p *Post) CanAddComments() bool {
	return p.CommentsEnabled
}
