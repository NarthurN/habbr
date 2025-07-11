package subscription

import (
	"context"
	"sync"
	"time"

	"github.com/NarthurN/habbr/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Subscriber представляет подписчика
type Subscriber struct {
	ID        string
	PostID    uuid.UUID
	Channel   chan *model.CommentSubscriptionPayload
	CreatedAt time.Time
	LastSeen  time.Time
}

// SubscriptionMetrics содержит метрики подписок
type SubscriptionMetrics struct {
	TotalSubscribers   int
	ActiveConnections  map[uuid.UUID]int // postID -> count
	MessagesSent       int64
	MessagesDropped    int64
	SubscriptionsTotal int64
}

// Service реализует сервис подписок с pub/sub системой
type Service struct {
	mu              sync.RWMutex
	subscribers     map[uuid.UUID]map[string]*Subscriber // postID -> subscriberID -> subscriber
	logger          *zap.Logger
	metrics         *SubscriptionMetrics
	channelSize     int
	cleanupInterval time.Duration
	maxIdleTime     time.Duration
}

// NewService создает новый сервис подписок
func NewService(logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}

	service := &Service{
		subscribers:     make(map[uuid.UUID]map[string]*Subscriber),
		logger:          logger,
		channelSize:     100,              // размер буфера канала
		cleanupInterval: 30 * time.Minute, // интервал очистки неактивных соединений
		maxIdleTime:     60 * time.Minute, // максимальное время бездействия
		metrics: &SubscriptionMetrics{
			ActiveConnections: make(map[uuid.UUID]int),
		},
	}

	// Запуск фоновой очистки
	go service.startCleanupRoutine()

	logger.Info("Subscription service initialized",
		zap.Int("channel_size", service.channelSize),
		zap.Duration("cleanup_interval", service.cleanupInterval),
		zap.Duration("max_idle_time", service.maxIdleTime),
	)

	return service
}

// SubscribeToComments создает подписку на комментарии к посту
func (s *Service) SubscribeToComments(ctx context.Context, postID uuid.UUID) (<-chan *model.CommentSubscriptionPayload, error) {
	if postID == uuid.Nil {
		s.logger.Warn("Attempt to subscribe with nil post ID")
		return nil, model.NewValidationError("post_id", "post ID is required")
	}

	s.logger.Debug("Creating new subscription", zap.String("post_id", postID.String()))

	// Создание канала для подписчика
	channel := make(chan *model.CommentSubscriptionPayload, s.channelSize)

	// Генерация уникального ID подписчика
	subscriberID := uuid.New().String()

	// Создание подписчика
	now := time.Now()
	subscriber := &Subscriber{
		ID:        subscriberID,
		PostID:    postID,
		Channel:   channel,
		CreatedAt: now,
		LastSeen:  now,
	}

	s.mu.Lock()
	// Инициализация карты подписчиков для поста, если она не существует
	if s.subscribers[postID] == nil {
		s.subscribers[postID] = make(map[string]*Subscriber)
	}

	// Добавление подписчика
	s.subscribers[postID][subscriberID] = subscriber

	// Обновление метрик
	s.metrics.TotalSubscribers++
	s.metrics.ActiveConnections[postID] = len(s.subscribers[postID])
	s.metrics.SubscriptionsTotal++
	s.mu.Unlock()

	s.logger.Info("Subscription created successfully",
		zap.String("post_id", postID.String()),
		zap.String("subscriber_id", subscriberID),
		zap.Int("post_subscribers", s.metrics.ActiveConnections[postID]),
		zap.Int("total_subscribers", s.metrics.TotalSubscribers),
	)

	// Запуск горутины для мониторинга контекста
	go s.monitorContext(ctx, postID, subscriberID)

	return channel, nil
}

// Unsubscribe отписывает от комментариев
func (s *Service) Unsubscribe(ctx context.Context, postID uuid.UUID, subscriberID string) error {
	s.logger.Debug("Unsubscribing",
		zap.String("post_id", postID.String()),
		zap.String("subscriber_id", subscriberID),
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверка существования подписок для поста
	postSubscribers, exists := s.subscribers[postID]
	if !exists {
		s.logger.Debug("No subscribers found for post", zap.String("post_id", postID.String()))
		return nil // подписка уже не существует
	}

	// Получение подписчика
	subscriber, exists := postSubscribers[subscriberID]
	if !exists {
		s.logger.Debug("Subscriber not found",
			zap.String("post_id", postID.String()),
			zap.String("subscriber_id", subscriberID),
		)
		return nil // подписчик уже не существует
	}

	// Закрытие канала и удаление подписчика
	s.safeCloseChannel(subscriber.Channel)
	delete(postSubscribers, subscriberID)

	// Обновление метрик
	s.metrics.TotalSubscribers--
	if len(postSubscribers) == 0 {
		delete(s.subscribers, postID)
		delete(s.metrics.ActiveConnections, postID)
	} else {
		s.metrics.ActiveConnections[postID] = len(postSubscribers)
	}

	s.logger.Info("Subscription removed successfully",
		zap.String("post_id", postID.String()),
		zap.String("subscriber_id", subscriberID),
		zap.Duration("subscription_duration", time.Since(subscriber.CreatedAt)),
		zap.Int("remaining_post_subscribers", len(postSubscribers)),
	)

	return nil
}

// NotifyCommentCreated уведомляет о создании нового комментария
func (s *Service) NotifyCommentCreated(ctx context.Context, comment *model.Comment) error {
	s.logger.Debug("Notifying comment created",
		zap.String("comment_id", comment.ID.String()),
		zap.String("post_id", comment.PostID.String()),
	)

	payload := &model.CommentSubscriptionPayload{
		PostID:     comment.PostID,
		Comment:    comment,
		ActionType: "CREATED",
	}

	return s.notifySubscribers(comment.PostID, payload)
}

// NotifyCommentUpdated уведомляет об обновлении комментария
func (s *Service) NotifyCommentUpdated(ctx context.Context, comment *model.Comment) error {
	s.logger.Debug("Notifying comment updated",
		zap.String("comment_id", comment.ID.String()),
		zap.String("post_id", comment.PostID.String()),
	)

	payload := &model.CommentSubscriptionPayload{
		PostID:     comment.PostID,
		Comment:    comment,
		ActionType: "UPDATED",
	}

	return s.notifySubscribers(comment.PostID, payload)
}

// NotifyCommentDeleted уведомляет об удалении комментария
func (s *Service) NotifyCommentDeleted(ctx context.Context, postID uuid.UUID, commentID uuid.UUID) error {
	s.logger.Debug("Notifying comment deleted",
		zap.String("comment_id", commentID.String()),
		zap.String("post_id", postID.String()),
	)

	payload := &model.CommentSubscriptionPayload{
		PostID: postID,
		Comment: &model.Comment{
			ID:     commentID,
			PostID: postID,
		},
		ActionType: "DELETED",
	}

	return s.notifySubscribers(postID, payload)
}

// notifySubscribers отправляет уведомление всем подписчикам поста
func (s *Service) notifySubscribers(postID uuid.UUID, payload *model.CommentSubscriptionPayload) error {
	s.mu.RLock()

	// Получение подписчиков для поста
	postSubscribers, exists := s.subscribers[postID]
	if !exists {
		s.mu.RUnlock()
		s.logger.Debug("No subscribers to notify", zap.String("post_id", postID.String()))
		return nil // нет подписчиков
	}

	// Создаем копию списка подписчиков для безопасной итерации
	subscribersCopy := make([]*Subscriber, 0, len(postSubscribers))
	for _, subscriber := range postSubscribers {
		subscribersCopy = append(subscribersCopy, subscriber)
	}
	s.mu.RUnlock()

	// Отправка уведомления всем подписчикам
	sentCount := 0
	droppedCount := 0

	for _, subscriber := range subscribersCopy {
		select {
		case subscriber.Channel <- payload:
			sentCount++
			// Обновляем время последней активности
			s.mu.Lock()
			if sub, exists := s.subscribers[postID][subscriber.ID]; exists {
				sub.LastSeen = time.Now()
			}
			s.mu.Unlock()
		default:
			// Канал заблокирован или закрыт, пропускаем
			droppedCount++
			s.logger.Warn("Message dropped for subscriber",
				zap.String("subscriber_id", subscriber.ID),
				zap.String("post_id", postID.String()),
				zap.String("action", payload.ActionType),
			)
		}
	}

	// Обновление метрик
	s.mu.Lock()
	s.metrics.MessagesSent += int64(sentCount)
	s.metrics.MessagesDropped += int64(droppedCount)
	s.mu.Unlock()

	s.logger.Debug("Notification sent to subscribers",
		zap.String("post_id", postID.String()),
		zap.String("action", payload.ActionType),
		zap.Int("sent", sentCount),
		zap.Int("dropped", droppedCount),
	)

	return nil
}

// monitorContext отслеживает отмену контекста и автоматически отписывает
func (s *Service) monitorContext(ctx context.Context, postID uuid.UUID, subscriberID string) {
	<-ctx.Done()

	s.logger.Debug("Context cancelled, unsubscribing",
		zap.String("post_id", postID.String()),
		zap.String("subscriber_id", subscriberID),
		zap.Error(ctx.Err()),
	)

	// Автоматическая отписка при отмене контекста
	if err := s.Unsubscribe(context.Background(), postID, subscriberID); err != nil {
		s.logger.Warn("Failed to unsubscribe on context cancellation",
			zap.Error(err),
			zap.String("post_id", postID.String()),
			zap.String("subscriber_id", subscriberID),
		)
	}
}

// startCleanupRoutine запускает фоновую очистку неактивных соединений
func (s *Service) startCleanupRoutine() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupIdleSubscribers()
	}
}

// cleanupIdleSubscribers удаляет неактивные подписки
func (s *Service) cleanupIdleSubscribers() {
	s.logger.Debug("Starting cleanup of idle subscribers")

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cleanedCount := 0

	for postID, postSubscribers := range s.subscribers {
		for subscriberID, subscriber := range postSubscribers {
			if now.Sub(subscriber.LastSeen) > s.maxIdleTime {
				s.logger.Debug("Cleaning up idle subscriber",
					zap.String("subscriber_id", subscriberID),
					zap.String("post_id", postID.String()),
					zap.Duration("idle_time", now.Sub(subscriber.LastSeen)),
				)

				// Закрытие канала
				s.safeCloseChannel(subscriber.Channel)
				delete(postSubscribers, subscriberID)
				cleanedCount++
			}
		}

		// Удаление пустых карт
		if len(postSubscribers) == 0 {
			delete(s.subscribers, postID)
			delete(s.metrics.ActiveConnections, postID)
		} else {
			s.metrics.ActiveConnections[postID] = len(postSubscribers)
		}
	}

	s.metrics.TotalSubscribers -= cleanedCount

	if cleanedCount > 0 {
		s.logger.Info("Cleanup completed",
			zap.Int("cleaned_subscribers", cleanedCount),
			zap.Int("remaining_subscribers", s.metrics.TotalSubscribers),
		)
	}
}

// safeCloseChannel безопасно закрывает канал
func (s *Service) safeCloseChannel(ch chan *model.CommentSubscriptionPayload) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Warn("Recovered from panic while closing channel", zap.Any("panic", r))
		}
	}()

	select {
	case <-ch:
		// Канал уже закрыт
	default:
		close(ch)
	}
}

// GetSubscriberCount возвращает количество подписчиков для поста
func (s *Service) GetSubscriberCount(postID uuid.UUID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	postSubscribers, exists := s.subscribers[postID]
	if !exists {
		return 0
	}

	return len(postSubscribers)
}

// GetTotalSubscriberCount возвращает общее количество подписчиков
func (s *Service) GetTotalSubscriberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.metrics.TotalSubscribers
}

// GetMetrics возвращает метрики подписок
func (s *Service) GetMetrics() SubscriptionMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Создаем копию для безопасного возврата
	metrics := SubscriptionMetrics{
		TotalSubscribers:   s.metrics.TotalSubscribers,
		ActiveConnections:  make(map[uuid.UUID]int),
		MessagesSent:       s.metrics.MessagesSent,
		MessagesDropped:    s.metrics.MessagesDropped,
		SubscriptionsTotal: s.metrics.SubscriptionsTotal,
	}

	for postID, count := range s.metrics.ActiveConnections {
		metrics.ActiveConnections[postID] = count
	}

	return metrics
}

// HealthCheck проверяет состояние сервис подписок
func (s *Service) HealthCheck() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalSubs := 0
	for _, postSubs := range s.subscribers {
		totalSubs += len(postSubs)
	}

	if totalSubs != s.metrics.TotalSubscribers {
		s.logger.Error("Metrics mismatch detected",
			zap.Int("actual_subscribers", totalSubs),
			zap.Int("metric_subscribers", s.metrics.TotalSubscribers),
		)
		return model.NewInternalError("subscription service metrics inconsistency")
	}

	return nil
}

// Close закрывает все подписки и очищает ресурсы
func (s *Service) Close() {
	s.logger.Info("Shutting down subscription service")

	s.mu.Lock()
	defer s.mu.Unlock()

	totalClosed := 0

	// Закрытие всех каналов
	for postID, postSubscribers := range s.subscribers {
		for subscriberID, subscriber := range postSubscribers {
			s.safeCloseChannel(subscriber.Channel)
			delete(postSubscribers, subscriberID)
			totalClosed++
		}
		delete(s.subscribers, postID)
	}

	// Очистка метрик
	s.metrics.TotalSubscribers = 0
	s.metrics.ActiveConnections = make(map[uuid.UUID]int)

	s.logger.Info("Subscription service shutdown completed",
		zap.Int("closed_subscriptions", totalClosed),
		zap.Int64("total_messages_sent", s.metrics.MessagesSent),
		zap.Int64("total_messages_dropped", s.metrics.MessagesDropped),
		zap.Int64("total_subscriptions_created", s.metrics.SubscriptionsTotal),
	)
}

// Subscribe создает новую подписку на события комментариев для указанного поста
func (s *Service) Subscribe(ctx context.Context, postID uuid.UUID) (<-chan *model.CommentSubscriptionPayload, error) {
	return s.SubscribeToComments(ctx, postID)
}

// Publish отправляет событие всем подписчикам указанного поста
func (s *Service) Publish(postID uuid.UUID, payload *model.CommentSubscriptionPayload) {
	if err := s.notifySubscribers(postID, payload); err != nil {
		s.logger.Warn("Failed to publish message to subscribers",
			zap.Error(err),
			zap.String("post_id", postID.String()),
			zap.String("action_type", payload.ActionType),
		)
	}
}

// Shutdown корректно завершает работу сервиса подписок
func (s *Service) Shutdown() {
	s.Close()
}
