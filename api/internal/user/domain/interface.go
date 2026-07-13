package domain

import (
	"context"

	"alexdenkk/labs/pkg/token/jwt"

	"github.com/google/uuid"
)

type UserService interface {
	SignUp(context.Context, string, string) error
	SignIn(context.Context, string, string) (string, string, error)
	RefreshToken(context.Context, string) (string, string, error)
	GetSelf(context.Context, *jwt.Claims) (User, error)
}

type UserRepository interface {
	Get(context.Context, uuid.UUID) (User, error)
	GetByEmail(context.Context, string) (User, error)
	Create(context.Context, User) error
	Update(context.Context, User) error
}
