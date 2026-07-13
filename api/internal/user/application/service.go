package application

import (
	"alexdenkk/labs/internal/user/domain"
	"alexdenkk/labs/pkg/config"
	"alexdenkk/labs/pkg/token/jwt"
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	repository   domain.UserRepository
	agentConfig  *config.AgentConfig
	tokenManager *jwt.TokenManager
}

func New(
	agentCfg *config.AgentConfig,
	tokenManager *jwt.TokenManager,
	repository domain.UserRepository,
) domain.UserService {
	return &service{
		repository:   repository,
		agentConfig:  agentCfg,
		tokenManager: tokenManager,
	}
}

// Авторизация пользователя по email и паролю
func (service *service) SignIn(
	ctx context.Context, email, password string,
) (string, string, error) {
	// Вызов репозитория для получения записи
	user, err := service.repository.GetByEmail(ctx, email)

	if err != nil {
		return "", "", errors.New("invalid email or password")
	}

	// Сравнение хэшей пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))

	if err != nil {
		return "", "", errors.New("invalid email or password")
	}

	// Генерация токенов авторизации
	accessToken, refreshToken, err := service.tokenManager.GenerateTokenPair(user.ID, user.Email)

	if err != nil {
		return "", "", errors.New("error while authorizing")
	}

	return accessToken, refreshToken, nil
}

// Регистрация пользователя
func (service *service) SignUp(
	ctx context.Context, email, password string,
) error {
	// Вызов репозитория
	_, err := service.repository.GetByEmail(ctx, email)

	if err == nil {
		return errors.New("user with this email already exists")
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return errors.New("error while registering user")
	}

	// Запись пользователя
	user := domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	user.ID = uuid.New()

	err = user.Validate()

	if err != nil {
		return err
	}

	// Вызов репозитория
	err = service.repository.Create(ctx, user)

	if err != nil {
		return errors.New("error while registering user")
	}

	return nil
}

func (service *service) GetSelf(ctx context.Context, claims *jwt.Claims) (domain.User, error) {
	user, err := service.repository.Get(ctx, claims.UserID)

	if err != nil {
		return domain.User{}, errors.New("user not found")
	}

	return user, nil
}

// Обновление токена
func (service *service) RefreshToken(
	ctx context.Context, refreshToken string,
) (string, string, error) {
	// Парсинг токена
	id, err := service.tokenManager.ParseRefreshToken(refreshToken)

	if err != nil {
		return "", "", errors.New("error while generating token")
	}

	// Вызов репозитория
	user, err := service.repository.Get(ctx, id)

	if err != nil {
		return "", "", errors.New("user not found")
	}

	// Генерация токена доступа
	accessToken, err := service.tokenManager.GenerateAccessToken(user.ID, user.Email)

	if err != nil {
		return "", "", errors.New("error while generating token")
	}

	// Генерация refresh токена
	newRefreshToken, err := service.tokenManager.GenerateRefreshToken(user.ID)

	if err != nil {
		return "", "", errors.New("error while generating token")
	}

	return accessToken, newRefreshToken, nil
}
