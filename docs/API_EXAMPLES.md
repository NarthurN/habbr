# API Examples - Habbr GraphQL

Этот документ содержит практические примеры использования GraphQL API Habbr с детальными описаниями и ответами.

## Оглавление

- [Основы работы с API](#основы-работы-с-api)
- [Queries (Запросы)](#queries-запросы)
- [Mutations (Мутации)](#mutations-мутации)
- [Subscriptions (Подписки)](#subscriptions-подписки)
- [Обработка ошибок](#обработка-ошибок)
- [Паттерны и лучшие практики](#паттерны-и-лучшие-практики)

## Основы работы с API

### Endpoint

```
POST http://localhost:8080/query
Content-Type: application/json
```

### GraphQL Playground

В режиме разработки доступен по адресу: `http://localhost:8080/`

### Аутентификация

```javascript
// Заголовки запроса (пример)
{
  "Authorization": "Bearer <your-jwt-token>",
  "Content-Type": "application/json"
}
```

## Queries (Запросы)

### 1. Получение списка постов

#### Базовый запрос

```graphql
query GetPosts {
  posts(first: 10) {
    edges {
      node {
        id
        title
        content
        authorID
        commentsEnabled
        createdAt
        updatedAt
      }
      cursor
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    totalCount
  }
}
```

#### Ответ

```json
{
  "data": {
    "posts": {
      "edges": [
        {
          "node": {
            "id": "550e8400-e29b-41d4-a716-446655440001",
            "title": "Welcome to Habbr",
            "content": "This is our first post. Feel free to comment!",
            "authorID": "550e8400-e29b-41d4-a716-446655440000",
            "commentsEnabled": true,
            "createdAt": "2023-12-01T10:00:00Z",
            "updatedAt": "2023-12-01T10:00:00Z"
          },
          "cursor": "WzIwMjMtMTItMDFUMTA6MDA6MDBaXQ=="
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "WzIwMjMtMTItMDFUMTA6MDA6MDBaXQ==",
        "endCursor": "WzIwMjMtMTItMDFUMTA6MDA6MDBaXQ=="
      },
      "totalCount": 1
    }
  }
}
```

#### С пагинацией

```graphql
query GetPostsWithPagination($first: Int!, $after: String) {
  posts(first: $first, after: $after) {
    edges {
      node {
        id
        title
        authorID
        createdAt
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

**Переменные:**
```json
{
  "first": 5,
  "after": "WzIwMjMtMTItMDFUMTA6MDA6MDBaXQ=="
}
```

#### С фильтрацией

```graphql
query GetFilteredPosts($filter: PostFilter!) {
  posts(first: 10, filter: $filter) {
    edges {
      node {
        id
        title
        authorID
        commentsEnabled
      }
    }
    totalCount
  }
}
```

**Переменные:**
```json
{
  "filter": {
    "authorID": "550e8400-e29b-41d4-a716-446655440000",
    "commentsEnabled": true
  }
}
```

### 2. Получение конкретного поста

#### С комментариями

```graphql
query GetPost($id: ID!) {
  post(id: $id) {
    id
    title
    content
    authorID
    commentsEnabled
    createdAt
    updatedAt
    comments(first: 10) {
      edges {
        node {
          id
          content
          authorID
          depth
          createdAt
          children(first: 5) {
            edges {
              node {
                id
                content
                authorID
                depth
              }
            }
          }
        }
      }
    }
  }
}
```

**Переменные:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001"
}
```

#### Ответ

```json
{
  "data": {
    "post": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "title": "Welcome to Habbr",
      "content": "This is our first post. Feel free to comment!",
      "authorID": "550e8400-e29b-41d4-a716-446655440000",
      "commentsEnabled": true,
      "createdAt": "2023-12-01T10:00:00Z",
      "updatedAt": "2023-12-01T10:00:00Z",
      "comments": {
        "edges": [
          {
            "node": {
              "id": "550e8400-e29b-41d4-a716-446655440010",
              "content": "Great post! Looking forward to more.",
              "authorID": "550e8400-e29b-41d4-a716-446655440002",
              "depth": 0,
              "createdAt": "2023-12-01T11:00:00Z",
              "children": {
                "edges": [
                  {
                    "node": {
                      "id": "550e8400-e29b-41d4-a716-446655440011",
                      "content": "I agree!",
                      "authorID": "550e8400-e29b-41d4-a716-446655440003",
                      "depth": 1
                    }
                  }
                ]
              }
            }
          }
        ]
      }
    }
  }
}
```

### 3. Получение дерева комментариев

```graphql
query GetCommentsTree($postID: ID!) {
  commentsTree(postID: $postID) {
    id
    content
    authorID
    depth
    parentID
    createdAt
    children {
      id
      content
      authorID
      depth
      children {
        id
        content
        depth
      }
    }
  }
}
```

### 4. Поиск постов

```graphql
query SearchPosts($query: String!, $first: Int) {
  searchPosts(query: $query, first: $first) {
    edges {
      node {
        id
        title
        content
        authorID
        createdAt
      }
    }
    totalCount
  }
}
```

**Переменные:**
```json
{
  "query": "GraphQL API",
  "first": 10
}
```

### 5. Статистика постов

```graphql
query GetPostStats($postID: ID!) {
  postStats(postID: $postID) {
    totalComments
    maxCommentDepth
    averageCommentDepth
    lastCommentAt
  }
}
```

## Mutations (Мутации)

### 1. Создание поста

```graphql
mutation CreatePost($input: PostInput!) {
  createPost(input: $input) {
    success
    post {
      id
      title
      content
      authorID
      commentsEnabled
      createdAt
    }
    error
  }
}
```

**Переменные:**
```json
{
  "input": {
    "title": "My New Post",
    "content": "This is the content of my new post. It can be quite long and contain various information.",
    "authorID": "550e8400-e29b-41d4-a716-446655440000",
    "commentsEnabled": true
  }
}
```

#### Ответ (успех)

```json
{
  "data": {
    "createPost": {
      "success": true,
      "post": {
        "id": "550e8400-e29b-41d4-a716-446655440005",
        "title": "My New Post",
        "content": "This is the content of my new post...",
        "authorID": "550e8400-e29b-41d4-a716-446655440000",
        "commentsEnabled": true,
        "createdAt": "2023-12-01T15:30:00Z"
      },
      "error": null
    }
  }
}
```

#### Ответ (ошибка)

```json
{
  "data": {
    "createPost": {
      "success": false,
      "post": null,
      "error": "Title cannot be empty"
    }
  }
}
```

### 2. Обновление поста

```graphql
mutation UpdatePost($id: ID!, $input: PostUpdateInput!) {
  updatePost(id: $id, input: $input) {
    success
    post {
      id
      title
      content
      commentsEnabled
      updatedAt
    }
    error
  }
}
```

**Переменные:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "input": {
    "title": "Updated Post Title",
    "content": "Updated content here...",
    "commentsEnabled": false
  }
}
```

### 3. Удаление поста

```graphql
mutation DeletePost($id: ID!) {
  deletePost(id: $id) {
    success
    deletedID
    error
  }
}
```

**Переменные:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001"
}
```

### 4. Создание комментария

```graphql
mutation CreateComment($input: CommentInput!) {
  createComment(input: $input) {
    success
    comment {
      id
      content
      authorID
      postID
      parentID
      depth
      createdAt
    }
    error
  }
}
```

**Переменные (корневой комментарий):**
```json
{
  "input": {
    "postID": "550e8400-e29b-41d4-a716-446655440001",
    "content": "This is a great post! Thanks for sharing.",
    "authorID": "550e8400-e29b-41d4-a716-446655440002"
  }
}
```

**Переменные (ответ на комментарий):**
```json
{
  "input": {
    "postID": "550e8400-e29b-41d4-a716-446655440001",
    "parentID": "550e8400-e29b-41d4-a716-446655440010",
    "content": "I totally agree with your comment!",
    "authorID": "550e8400-e29b-41d4-a716-446655440003"
  }
}
```

### 5. Обновление комментария

```graphql
mutation UpdateComment($id: ID!, $input: CommentUpdateInput!) {
  updateComment(id: $id, input: $input) {
    success
    comment {
      id
      content
      updatedAt
    }
    error
  }
}
```

### 6. Удаление комментария

```graphql
mutation DeleteComment($id: ID!) {
  deleteComment(id: $id) {
    success
    deletedID
    error
  }
}
```

### 7. Пакетные операции

#### Создание нескольких постов

```graphql
mutation CreateMultiplePosts($inputs: [PostInput!]!) {
  createPosts(inputs: $inputs) {
    success
    createdCount
    posts {
      id
      title
      authorID
    }
    errors
  }
}
```

**Переменные:**
```json
{
  "inputs": [
    {
      "title": "First Post",
      "content": "Content of first post",
      "authorID": "550e8400-e29b-41d4-a716-446655440000",
      "commentsEnabled": true
    },
    {
      "title": "Second Post",
      "content": "Content of second post",
      "authorID": "550e8400-e29b-41d4-a716-446655440000",
      "commentsEnabled": true
    }
  ]
}
```

#### Удаление нескольких комментариев

```graphql
mutation DeleteMultipleComments($ids: [ID!]!) {
  deleteComments(ids: $ids) {
    success
    deletedCount
    deletedIDs
    errors
  }
}
```

### 8. Управление комментариями к посту

```graphql
mutation TogglePostComments($postID: ID!, $enabled: Boolean!) {
  togglePostComments(postID: $postID, enabled: $enabled) {
    success
    post {
      id
      commentsEnabled
    }
    error
  }
}
```

## Subscriptions (Подписки)

### 1. Подписка на события комментариев

```graphql
subscription CommentEvents($postID: ID!) {
  commentEvents(postID: $postID) {
    type
    comment {
      id
      content
      authorID
      depth
      createdAt
    }
    postID
  }
}
```

**Переменные:**
```json
{
  "postID": "550e8400-e29b-41d4-a716-446655440001"
}
```

#### Пример событий

**Создание комментария:**
```json
{
  "data": {
    "commentEvents": {
      "type": "CREATED",
      "comment": {
        "id": "550e8400-e29b-41d4-a716-446655440020",
        "content": "New comment here!",
        "authorID": "550e8400-e29b-41d4-a716-446655440004",
        "depth": 0,
        "createdAt": "2023-12-01T16:00:00Z"
      },
      "postID": "550e8400-e29b-41d4-a716-446655440001"
    }
  }
}
```

**Обновление комментария:**
```json
{
  "data": {
    "commentEvents": {
      "type": "UPDATED",
      "comment": {
        "id": "550e8400-e29b-41d4-a716-446655440020",
        "content": "Updated comment content",
        "authorID": "550e8400-e29b-41d4-a716-446655440004",
        "depth": 0,
        "createdAt": "2023-12-01T16:00:00Z"
      },
      "postID": "550e8400-e29b-41d4-a716-446655440001"
    }
  }
}
```

### 2. Подписка на статистику постов

```graphql
subscription PostStatistics($postID: ID!) {
  postStatistics(postID: $postID) {
    totalComments
    maxCommentDepth
    lastCommentAt
  }
}
```

### 3. JavaScript клиент (пример)

```javascript
import { createClient } from 'graphql-ws';

const client = createClient({
  url: 'ws://localhost:8080/query',
});

// Подписка на комментарии
const unsubscribe = client.subscribe(
  {
    query: `
      subscription CommentEvents($postID: ID!) {
        commentEvents(postID: $postID) {
          type
          comment {
            id
            content
            authorID
            depth
          }
        }
      }
    `,
    variables: { postID: '550e8400-e29b-41d4-a716-446655440001' }
  },
  {
    next: (data) => {
      console.log('New comment event:', data);
    },
    error: (err) => {
      console.error('Subscription error:', err);
    },
    complete: () => {
      console.log('Subscription completed');
    }
  }
);

// Отписка
// unsubscribe();
```

## Обработка ошибок

### 1. Ошибки валидации

```json
{
  "data": {
    "createPost": {
      "success": false,
      "post": null,
      "error": "Title must be between 1 and 200 characters"
    }
  }
}
```

### 2. Ошибки не найдено

```json
{
  "data": {
    "post": null
  },
  "errors": [
    {
      "message": "Post not found",
      "locations": [{"line": 2, "column": 3}],
      "path": ["post"],
      "extensions": {
        "code": "NOT_FOUND",
        "postID": "550e8400-e29b-41d4-a716-446655440999"
      }
    }
  ]
}
```

### 3. Ошибки сети/системы

```json
{
  "errors": [
    {
      "message": "Database connection failed",
      "extensions": {
        "code": "INTERNAL_ERROR"
      }
    }
  ]
}
```

## Паттерны и лучшие практики

### 1. Фрагменты для переиспользования

```graphql
fragment PostBasicInfo on Post {
  id
  title
  authorID
  createdAt
  commentsEnabled
}

fragment CommentInfo on Comment {
  id
  content
  authorID
  depth
  createdAt
}

query GetPostsWithFragments {
  posts(first: 10) {
    edges {
      node {
        ...PostBasicInfo
        comments(first: 3) {
          edges {
            node {
              ...CommentInfo
            }
          }
        }
      }
    }
  }
}
```

### 2. Условные поля

```graphql
query GetPost($id: ID!, $includeComments: Boolean!) {
  post(id: $id) {
    id
    title
    content
    comments(first: 10) @include(if: $includeComments) {
      edges {
        node {
          id
          content
        }
      }
    }
  }
}
```

### 3. Алиасы для нескольких запросов

```graphql
query GetMultiplePosts {
  latestPosts: posts(first: 5) {
    edges {
      node {
        id
        title
      }
    }
  }

  popularPosts: posts(first: 5, filter: { sortBy: COMMENTS_COUNT }) {
    edges {
      node {
        id
        title
      }
    }
  }
}
```

### 4. Переменные по умолчанию

```graphql
query GetPosts($first: Int = 10, $includeContent: Boolean = false) {
  posts(first: $first) {
    edges {
      node {
        id
        title
        content @include(if: $includeContent)
      }
    }
  }
}
```

### 5. Обработка loading состояний

```javascript
// React пример с Apollo Client
const { loading, error, data } = useQuery(GET_POSTS, {
  variables: { first: 10 },
  errorPolicy: 'partial'
});

if (loading) return <LoadingSpinner />;
if (error) return <ErrorMessage error={error} />;

return (
  <PostList posts={data.posts.edges.map(edge => edge.node)} />
);
```

### 6. Оптимистические обновления

```javascript
const [createPost] = useMutation(CREATE_POST, {
  optimisticResponse: {
    createPost: {
      __typename: 'PostResult',
      success: true,
      post: {
        __typename: 'Post',
        id: 'temp-id',
        title: variables.input.title,
        content: variables.input.content,
        authorID: variables.input.authorID,
        commentsEnabled: variables.input.commentsEnabled,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      },
      error: null
    }
  },
  update: (cache, { data }) => {
    // Обновление кэша
  }
});
```

### 7. Пагинация с cursor

```javascript
const { data, fetchMore } = useQuery(GET_POSTS, {
  variables: { first: 10 }
});

const loadMore = () => {
  fetchMore({
    variables: {
      after: data.posts.pageInfo.endCursor
    },
    updateQuery: (prev, { fetchMoreResult }) => {
      if (!fetchMoreResult) return prev;

      return {
        posts: {
          ...fetchMoreResult.posts,
          edges: [...prev.posts.edges, ...fetchMoreResult.posts.edges]
        }
      };
    }
  });
};
```

## Заключение

Эти примеры покрывают основные случаи использования GraphQL API Habbr. Для получения полного списка доступных полей и типов используйте интроспекцию или GraphQL Playground.

Помните о следующих лучших практиках:
- Используйте фрагменты для переиспользования
- Запрашивайте только необходимые поля
- Используйте переменные вместо строковой интерполяции
- Обрабатывайте ошибки на уровне полей и операций
- Кэшируйте результаты где это возможно
