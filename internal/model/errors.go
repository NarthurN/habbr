package model

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Доменные ошибки
var (
	ErrPostNotFound     = errors.New("post not found")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrValidation       = errors.New("validation error")
	ErrCommentsDisabled = errors.New("comments are disabled for this post")
	ErrInvalidParent    = errors.New("invalid parent comment")
	ErrInternalError    = errors.New("internal server error")
)

// DomainError представляет доменную ошибку с дополнительным контекстом
type DomainError struct {
	Type    string            `json:"type"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewValidationError создает ошибку валидации
func NewValidationError(field, message string) *DomainError {
	return &DomainError{
		Type:    "VALIDATION_ERROR",
		Message: message,
		Details: map[string]string{
			"field": field,
		},
	}
}

// NewNotFoundError создает ошибку "не найдено"
func NewNotFoundError(entity string, id uuid.UUID) *DomainError {
	return &DomainError{
		Type:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", entity),
		Details: map[string]string{
			"entity": entity,
			"id":     id.String(),
		},
	}
}

// NewForbiddenError создает ошибку "запрещено"
func NewForbiddenError(action string) *DomainError {
	return &DomainError{
		Type:    "FORBIDDEN",
		Message: fmt.Sprintf("action '%s' is forbidden", action),
		Details: map[string]string{
			"action": action,
		},
	}
}

// NewUnauthorizedError создает ошибку "не авторизован"
func NewUnauthorizedError() *DomainError {
	return &DomainError{
		Type:    "UNAUTHORIZED",
		Message: "authentication required",
	}
}

// NewInternalError создает внутреннюю ошибку сервера
func NewInternalError(message string) *DomainError {
	return &DomainError{
		Type:    "INTERNAL_ERROR",
		Message: message,
	}
}
