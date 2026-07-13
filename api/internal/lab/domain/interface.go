package domain

import (
	"alexdenkk/labs/pkg/token/jwt"
	"context"

	"github.com/google/uuid"
)

type LabService interface {
	GetForSelf(context.Context, *jwt.Claims) ([]Lab, error)
	Get(context.Context, uuid.UUID, *jwt.Claims) (Lab, error)
	Generate(context.Context, Lab, *jwt.Claims) error
	Delete(context.Context, uuid.UUID, *jwt.Claims) error
}

type LabRepository interface {
	Get(context.Context, uuid.UUID) (Lab, error)
	GetByUser(context.Context, uuid.UUID) ([]Lab, error)
	Create(context.Context, Lab) error
	Delete(context.Context, uuid.UUID) error
}

type AgentAdapter interface {
	Call(context.Context, AgentRequest, string) (AgentResponse, error)
}
