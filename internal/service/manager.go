package service

import (
	"context"

	"github.com/NarthurN/habbr/internal/repository"
	"github.com/NarthurN/habbr/internal/service/comment"
	"github.com/NarthurN/habbr/internal/service/post"
	"github.com/NarthurN/habbr/internal/service/subscription"
	"go.uber.org/zap"
)

// Manager управляет всеми сервисами
type Manager struct {
	services *Services
	logger   *zap.Logger
}

// NewManager создает новый менеджер сервисов
func NewManager(repos *repository.Repositories, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}

	logger.Info("Initializing service manager")

	// Создаем сервис подписок
	subscriptionService := subscription.NewService(logger.Named("subscription"))

	// Создаем сервисы с dependency injection
	postService := post.NewService(repos, logger.Named("post"))
	commentService := comment.NewService(repos, logger.Named("comment"), subscriptionService)

	services := &Services{
		Post:         postService,
		Comment:      commentService,
		Subscription: subscriptionService,
	}

	logger.Info("Service manager initialized successfully")

	return &Manager{
		services: services,
		logger:   logger,
	}
}

// GetServices возвращает все сервисы
func (m *Manager) GetServices() *Services {
	return m.services
}

// HealthCheck проверяет состояние всех сервисов
func (m *Manager) HealthCheck(ctx context.Context) error {
	m.logger.Debug("Performing services health check")

	// Проверяем сервис подписок
	if subscriptionService, ok := m.services.Subscription.(*subscription.Service); ok {
		if err := subscriptionService.HealthCheck(); err != nil {
			m.logger.Error("Subscription service health check failed", zap.Error(err))
			return err
		}
	}

	m.logger.Debug("All services health check passed")
	return nil
}

// GetMetrics возвращает метрики всех сервисов
func (m *Manager) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Получаем метрики сервиса подписок
	if subscriptionService, ok := m.services.Subscription.(*subscription.Service); ok {
		metrics["subscription"] = subscriptionService.GetMetrics()
	}

	// Можно добавить метрики других сервисов в будущем

	return metrics
}

// Close закрывает все сервисы и освобождает ресурсы
func (m *Manager) Close() {
	m.logger.Info("Shutting down service manager")

	// Закрываем сервис подписок
	if subscriptionService, ok := m.services.Subscription.(*subscription.Service); ok {
		subscriptionService.Close()
	}

	m.logger.Info("Service manager shutdown completed")
}
