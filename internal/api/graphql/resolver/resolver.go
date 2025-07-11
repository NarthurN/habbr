package resolver

import (
	"github.com/NarthurN/habbr/internal/service"
	"go.uber.org/zap"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	services *service.Services
	logger   *zap.Logger
}

// NewResolver создает новый экземпляр резолвера с внедренными зависимостями
func NewResolver(services *service.Services, logger *zap.Logger) *Resolver {
	return &Resolver{
		services: services,
		logger:   logger,
	}
}
