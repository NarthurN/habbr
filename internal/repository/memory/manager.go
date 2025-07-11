package memory

import (
	"context"

	"github.com/NarthurN/habbr/internal/repository"
)

// Manager управляет in-memory репозиториями
type Manager struct {
	repositories *repository.Repositories
}

// NewManager создает новый менеджер in-memory репозиториев
func NewManager() *Manager {
	return &Manager{
		repositories: &repository.Repositories{
			Post:    NewPostRepository(),
			Comment: NewCommentRepository(),
		},
	}
}

// GetRepositories возвращает все репозитории
func (m *Manager) GetRepositories() *repository.Repositories {
	return m.repositories
}

// Close закрывает соединения (для in-memory реализации не требуется)
func (m *Manager) Close(ctx context.Context) error {
	return nil
}

// HealthCheck проверяет здоровье соединения (для in-memory всегда здоров)
func (m *Manager) HealthCheck(ctx context.Context) error {
	return nil
}

// Migrate выполняет миграции (для in-memory не требуется)
func (m *Manager) Migrate(ctx context.Context) error {
	return nil
}
