package model

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MaxCommentLength определяет максимальную длину содержимого комментария в символах.
// Ограничение помогает предотвратить злоупотребления и обеспечить разумный размер данных.
const MaxCommentLength = 2000

// Comment представляет доменную модель комментария в иерархической системе.
//
// Комментарии организованы в древовидную структуру с неограниченной глубиной вложенности.
// Каждый комментарий может иметь родительский комментарий (ParentID) и множество дочерних (Children).
// Корневые комментарии имеют ParentID = nil и привязаны непосредственно к посту.
//
// Поле Depth автоматически вычисляется при создании и показывает уровень вложенности:
// - 0 для корневых комментариев
// - 1 для ответов на корневые комментарии
// - 2 для ответов на ответы и так далее
//
// Пример использования:
//   comment := NewComment(CommentInput{
//       PostID: postID,
//       ParentID: &parentCommentID, // nil для корневого комментария
//       Content: "Содержимое комментария",
//       AuthorID: userID,
//   }, depth)
type Comment struct {
	// ID - уникальный идентификатор комментария в формате UUID
	ID uuid.UUID `json:"id"`

	// PostID - идентификатор поста, к которому относится комментарий
	PostID uuid.UUID `json:"post_id"`

	// ParentID - идентификатор родительского комментария (nil для корневых комментариев)
	ParentID *uuid.UUID `json:"parent_id"`

	// Content - текстовое содержимое комментария, максимум 2000 символов
	Content string `json:"content"`

	// AuthorID - идентификатор автора комментария
	AuthorID uuid.UUID `json:"author_id"`

	// Depth - глубина вложенности комментария в дереве (0 для корневых)
	Depth int `json:"depth"`

	// CreatedAt - время создания комментария
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt - время последнего обновления комментария
	UpdatedAt time.Time `json:"updated_at"`

	// Children - массив дочерних комментариев (заполняется при построении дерева)
	Children []*Comment `json:"children,omitempty"`
}

// CommentInput представляет входные данные для создания нового комментария.
//
// Все поля являются обязательными, кроме ParentID (для корневых комментариев).
// Валидация происходит через метод Validate().
//
// Пример использования:
//   // Корневой комментарий
//   input := CommentInput{
//       PostID: postID,
//       Content: "Мой комментарий к посту",
//       AuthorID: userID,
//   }
//
//   // Ответ на комментарий
//   reply := CommentInput{
//       PostID: postID,
//       ParentID: &parentCommentID,
//       Content: "Мой ответ на комментарий",
//       AuthorID: userID,
//   }
type CommentInput struct {
	// PostID - идентификатор поста, обязательное поле
	PostID uuid.UUID `json:"post_id"`

	// ParentID - идентификатор родительского комментария (опциональное для корневых комментариев)
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// Content - содержимое комментария, обязательное поле, от 1 до 2000 символов
	Content string `json:"content"`

	// AuthorID - идентификатор автора, обязательное поле
	AuthorID uuid.UUID `json:"author_id"`
}

// CommentUpdateInput представляет входные данные для обновления существующего комментария.
//
// В текущей реализации можно обновлять только содержимое комментария.
// Поле является опциональным (указатель), обновляется только если не равно nil.
//
// Пример использования:
//   newContent := "Обновленное содержимое комментария"
//   update := CommentUpdateInput{
//       Content: &newContent,
//   }
type CommentUpdateInput struct {
	// Content - новое содержимое комментария, опциональное поле
	Content *string `json:"content,omitempty"`
}

// CommentFilter представляет фильтры для поиска и выборки комментариев.
//
// Используется в операциях списочного получения комментариев для ограничения результатов.
// Все поля являются опциональными и могут комбинироваться.
//
// Пример использования:
//   // Все комментарии к посту
//   filter := CommentFilter{PostID: &postID}
//
//   // Только корневые комментарии
//   filter := CommentFilter{
//       PostID: &postID,
//       ParentID: nil, // специальное значение для корневых
//   }
//
//   // Комментарии определенного автора с ограничением глубины
//   maxDepth := 2
//   filter := CommentFilter{
//       PostID: &postID,
//       AuthorID: &userID,
//       MaxDepth: &maxDepth,
//   }
type CommentFilter struct {
	// PostID - фильтр по посту, если указан, возвращаются только комментарии к данному посту
	PostID *uuid.UUID `json:"post_id,omitempty"`

	// ParentID - фильтр по родительскому комментарию:
	// указанный UUID - только дочерние комментарии данного родителя
	// nil в фильтре - только корневые комментарии
	// не указан в фильтре - все комментарии независимо от иерархии
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// AuthorID - фильтр по автору, если указан, возвращаются только комментарии данного автора
	AuthorID *uuid.UUID `json:"author_id,omitempty"`

	// MaxDepth - максимальная глубина вложенности для включения в результат
	MaxDepth *int `json:"max_depth,omitempty"`
}

// CommentConnection представляет результат пагинированного запроса комментариев.
//
// Следует стандарту GraphQL Cursor Connections Specification.
// Содержит сами данные (edges) и метаинформацию о пагинации (pageInfo).
//
// Пример использования:
//   connection, err := commentService.ListComments(ctx, filter, pagination)
//   for _, edge := range connection.Edges {
//       comment := edge.Node
//       fmt.Printf("Comment by %s: %s\n", comment.AuthorID, comment.Content)
//   }
type CommentConnection struct {
	// Edges - массив ребер, каждое содержит комментарий и его cursor
	Edges []*CommentEdge `json:"edges"`

	// PageInfo - информация о пагинации и доступности соседних страниц
	PageInfo *PageInfo `json:"page_info"`
}

// CommentEdge представляет одно ребро в connection - комбинацию комментария и cursor.
//
// Cursor используется для получения следующих или предыдущих страниц.
// Это непрозрачная строка, которую не следует парсить на клиенте.
type CommentEdge struct {
	// Node - сам объект комментария
	Node *Comment `json:"node"`

	// Cursor - уникальный идентификатор позиции этого комментария в результатах
	Cursor string `json:"cursor"`
}

// CommentSubscriptionPayload представляет данные события для real-time подписок на комментарии.
//
// Используется в WebSocket подписках для уведомления клиентов о изменениях
// в комментариях к определенному посту в режиме реального времени.
//
// Пример использования:
//   // Подписка на события комментариев
//   subscription := `
//       subscription CommentEvents($postID: ID!) {
//           commentEvents(postID: $postID) {
//               postID
//               comment { id content authorID }
//               actionType
//           }
//       }
//   `
type CommentSubscriptionPayload struct {
	// PostID - идентификатор поста, к которому относится событие
	PostID uuid.UUID `json:"post_id"`

	// Comment - данные комментария (может быть nil для события удаления)
	Comment *Comment `json:"comment"`

	// ActionType - тип события: "CREATED", "UPDATED", "DELETED"
	ActionType string `json:"action_type"`
}

// Validate проверяет валидность входных данных для создания комментария.
//
// Выполняет следующие проверки:
// - Содержимое не пустое и не превышает MaxCommentLength символов
// - PostID не является пустым UUID
// - AuthorID не является пустым UUID
// - ParentID не проверяется (может быть nil для корневых комментариев)
//
// Возвращает:
//   - nil если все данные валидны
//   - error с описанием первой найденной ошибки валидации
//
// Пример использования:
//   input := CommentInput{
//       PostID: postID,
//       Content: "Содержимое комментария",
//       AuthorID: userID,
//   }
//   if err := input.Validate(); err != nil {
//       log.Printf("Ошибка валидации: %v", err)
//       return err
//   }
func (c *CommentInput) Validate() error {
	if strings.TrimSpace(c.Content) == "" {
		return errors.New("content cannot be empty")
	}

	if len(c.Content) > MaxCommentLength {
		return errors.New("content cannot exceed 2000 characters")
	}

	if c.PostID == uuid.Nil {
		return errors.New("post_id is required")
	}

	if c.AuthorID == uuid.Nil {
		return errors.New("author_id is required")
	}

	return nil
}

// Validate проверяет валидность данных для обновления комментария.
//
// Проверяет только поле Content, если оно указано (не равно nil):
// - Содержимое не пустое и не превышает MaxCommentLength символов
//
// Возвращает:
//   - nil если указанные данные валидны
//   - error с описанием первой найденной ошибки валидации
//
// Пример использования:
//   newContent := "Обновленное содержимое"
//   update := CommentUpdateInput{Content: &newContent}
//   if err := update.Validate(); err != nil {
//       return fmt.Errorf("ошибка валидации обновления: %w", err)
//   }
func (c *CommentUpdateInput) Validate() error {
	if c.Content != nil {
		if strings.TrimSpace(*c.Content) == "" {
			return errors.New("content cannot be empty")
		}

		if len(*c.Content) > MaxCommentLength {
			return errors.New("content cannot exceed 2000 characters")
		}
	}

	return nil
}

// NewComment создает новый комментарий из входных данных с автоматической генерацией ID и временных меток.
//
// Функция выполняет следующие действия:
// - Генерирует новый UUID для комментария
// - Обрезает пробелы в начале и конце содержимого
// - Устанавливает переданную глубину вложенности
// - Устанавливает текущее время как CreatedAt и UpdatedAt
// - Инициализирует пустой массив Children
// - Копирует остальные поля из входных данных
//
// Параметры:
//   - input: валидированные входные данные для создания комментария
//   - depth: глубина вложенности комментария в дереве (0 для корневых)
//
// Возвращает:
//   - *Comment: новый комментарий, готовый для сохранения в репозитории
//
// Примечание: Функция НЕ выполняет валидацию входных данных.
// Валидацию следует выполнить заранее через input.Validate().
//
// Пример использования:
//   input := CommentInput{
//       PostID: postID,
//       ParentID: &parentID,
//       Content: "  Содержимое с пробелами  ",
//       AuthorID: userID,
//   }
//   comment := NewComment(input, 1) // depth = 1 для ответа
//   // comment.Content теперь "Содержимое с пробелами" (без пробелов по краям)
func NewComment(input CommentInput, depth int) *Comment {
	now := time.Now()
	return &Comment{
		ID:        uuid.New(),
		PostID:    input.PostID,
		ParentID:  input.ParentID,
		Content:   strings.TrimSpace(input.Content),
		AuthorID:  input.AuthorID,
		Depth:     depth,
		CreatedAt: now,
		UpdatedAt: now,
		Children:  make([]*Comment, 0),
	}
}

// Update обновляет существующий комментарий новыми данными.
//
// Метод применяет изменения только к тем полям, которые указаны в input (не равны nil).
// В текущей реализации поддерживается обновление только содержимого.
// Автоматически обновляет временную метку UpdatedAt.
// Обрезает пробелы в начале и конце содержимого.
//
// Параметры:
//   - input: данные для обновления, поля равные nil игнорируются
//
// Побочные эффекты:
//   - Изменяет поле Content если оно указано в input
//   - Обновляет поле UpdatedAt на текущее время
//
// Примечание: Метод НЕ выполняет валидацию входных данных.
// Валидацию следует выполнить заранее через input.Validate().
//
// Пример использования:
//   newContent := "Обновленное содержимое комментария"
//   update := CommentUpdateInput{Content: &newContent}
//   comment.Update(update)
//   // comment.Content = "Обновленное содержимое комментария"
//   // comment.UpdatedAt = текущее время
func (c *Comment) Update(input CommentUpdateInput) {
	if input.Content != nil {
		c.Content = strings.TrimSpace(*input.Content)
	}

	c.UpdatedAt = time.Now()
}

// IsRootComment проверяет, является ли комментарий корневым (привязан непосредственно к посту).
//
// Корневые комментарии не имеют родительского комментария и представляют
// верхний уровень иерархии комментариев к посту.
//
// Возвращает:
//   - true если комментарий является корневым (ParentID == nil)
//   - false если комментарий является ответом на другой комментарий
//
// Пример использования:
//   if comment.IsRootComment() {
//       fmt.Println("Это корневой комментарий к посту")
//   } else {
//       fmt.Printf("Это ответ на комментарий %s\n", *comment.ParentID)
//   }
func (c *Comment) IsRootComment() bool {
	return c.ParentID == nil
}

// CanBeRepliedTo проверяет, можно ли создать ответ на данный комментарий.
//
// В текущей реализации всегда возвращает true, но метод оставлен для
// будущих возможных ограничений (например, максимальная глубина вложенности,
// архивирование старых комментариев, блокировка пользователей).
//
// Возвращает:
//   - true если на комментарий можно ответить
//   - false если ответы запрещены
//
// Пример использования:
//   if comment.CanBeRepliedTo() {
//       // Можно создавать ответ на комментарий
//       reply := NewComment(replyInput, comment.Depth + 1)
//   } else {
//       return errors.New("ответы на этот комментарий запрещены")
//   }
func (c *Comment) CanBeRepliedTo() bool {
	// Можно добавить ограничения на глубину вложенности если потребуется
	return true
}

// AddChild добавляет дочерний комментарий в коллекцию Children.
//
// Метод используется при построении древовидной структуры комментариев
// из плоского списка. Если массив Children не инициализирован, создает новый.
//
// Параметры:
//   - child: дочерний комментарий для добавления
//
// Побочные эффекты:
//   - Добавляет child в конец массива Children
//   - Инициализирует Children если он равен nil
//
// Примечание: Метод НЕ проверяет корректность связи parent-child.
// Вызывающий код должен убедиться, что child.ParentID == c.ID.
//
// Пример использования:
//   parent := &Comment{ID: parentID, Children: nil}
//   child := &Comment{ID: childID, ParentID: &parentID}
//   parent.AddChild(child)
//   // parent.Children теперь содержит child
func (c *Comment) AddChild(child *Comment) {
	if c.Children == nil {
		c.Children = make([]*Comment, 0)
	}
	c.Children = append(c.Children, child)
}

// GetDepth возвращает текущую глубину вложенности комментария в дереве.
//
// Глубина показывает уровень комментария в иерархии:
// - 0 для корневых комментариев (привязанных к посту)
// - 1 для ответов на корневые комментарии
// - 2 для ответов на ответы и так далее
//
// Возвращает:
//   - int: глубина вложенности (начиная с 0)
//
// Пример использования:
//   depth := comment.GetDepth()
//   if depth == 0 {
//       fmt.Println("Корневой комментарий")
//   } else {
//       fmt.Printf("Ответ уровня %d\n", depth)
//   }
func (c *Comment) GetDepth() int {
	return c.Depth
}

// BuildCommentsTree строит иерархическую древовидную структуру из плоского списка комментариев.
//
// Функция принимает неупорядоченный список комментариев и организует их в дерево,
// основываясь на связях ParentID. Корневые комментарии (ParentID == nil) становятся
// корнями деревьев, остальные размещаются под соответствующими родителями.
//
// Алгоритм:
// 1. Создает карту для быстрого поиска комментариев по ID
// 2. Инициализирует пустые массивы Children для всех комментариев
// 3. Проходит по всем комментариям и распределяет их по родителям
// 4. Возвращает только корневые комментарии (полные деревья)
//
// Параметры:
//   - comments: плоский список всех комментариев для построения дерева
//
// Возвращает:
//   - []*Comment: слайс корневых комментариев с заполненными Children
//
// Примечания:
// - Если родительский комментарий не найден в списке, дочерний комментарий игнорируется
// - Функция безопасна для пустого входного списка
// - Время выполнения O(n), где n - количество комментариев
//
// Пример использования:
//   flatComments := []*Comment{
//       {ID: id1, ParentID: nil},     // корневой
//       {ID: id2, ParentID: &id1},    // ответ на id1
//       {ID: id3, ParentID: &id1},    // еще один ответ на id1
//       {ID: id4, ParentID: &id2},    // ответ на id2
//   }
//   tree := BuildCommentsTree(flatComments)
//   // tree[0].Children содержит комментарии id2 и id3
//   // tree[0].Children[0].Children содержит комментарий id4
func BuildCommentsTree(comments []*Comment) []*Comment {
	if len(comments) == 0 {
		return make([]*Comment, 0)
	}

	// Создаем карту для быстрого поиска комментариев по ID
	commentMap := make(map[uuid.UUID]*Comment)
	for _, comment := range comments {
		comment.Children = make([]*Comment, 0) // Инициализируем дочерние комментарии
		commentMap[comment.ID] = comment
	}

	// Строим дерево
	rootComments := make([]*Comment, 0)
	for _, comment := range comments {
		if comment.IsRootComment() {
			rootComments = append(rootComments, comment)
		} else if parent, exists := commentMap[*comment.ParentID]; exists {
			parent.AddChild(comment)
		}
	}

	return rootComments
}

// FlattenCommentsTree преобразует иерархическую древовидную структуру в плоский упорядоченный список.
//
// Функция выполняет обход дерева в глубину (depth-first traversal) и собирает
// все комментарии в единый список, сохраняя иерархический порядок.
// Полезно для отображения комментариев в линейном виде с сохранением структуры.
//
// Алгоритм:
// - Рекурсивно обходит каждый узел дерева
// - Сначала добавляет родительский комментарий
// - Затем рекурсивно обрабатывает всех детей
// - Результат сохраняет порядок "родитель -> дети -> внуки"
//
// Параметры:
//   - tree: слайс корневых комментариев с заполненными Children
//
// Возвращает:
//   - []*Comment: плоский список всех комментариев в иерархическом порядке
//
// Примечания:
// - Функция безопасна для пустого входного дерева
// - Порядок следования: каждый родитель идет перед своими детьми
// - Время выполнения O(n), где n - общее количество комментариев
//
// Пример использования:
//   tree := BuildCommentsTree(comments)
//   flatList := FlattenCommentsTree(tree)
//
//   // Вывод комментариев с отступами по глубине
//   for _, comment := range flatList {
//       indent := strings.Repeat("  ", comment.GetDepth())
//       fmt.Printf("%s%s\n", indent, comment.Content)
//   }
//
//   // Результат:
//   // Корневой комментарий
//   //   Ответ на корневой
//   //     Ответ на ответ
//   //   Еще один ответ на корневой
func FlattenCommentsTree(tree []*Comment) []*Comment {
	result := make([]*Comment, 0)

	var flatten func([]*Comment)
	flatten = func(comments []*Comment) {
		for _, comment := range comments {
			result = append(result, comment)
			if len(comment.Children) > 0 {
				flatten(comment.Children)
			}
		}
	}

	flatten(tree)
	return result
}
