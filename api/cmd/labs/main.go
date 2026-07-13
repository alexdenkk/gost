package main

import (
	lab_app "alexdenkk/labs/internal/lab/application"
	lab_domain "alexdenkk/labs/internal/lab/domain"
	lab_agent "alexdenkk/labs/internal/lab/infrastructure/agent"
	lab_repository "alexdenkk/labs/internal/lab/infrastructure/postgres"
	lab_api "alexdenkk/labs/internal/lab/transport/api"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	user_app "alexdenkk/labs/internal/user/application"
	user_domain "alexdenkk/labs/internal/user/domain"
	user_repository "alexdenkk/labs/internal/user/infrastructure/postgres"
	user_api "alexdenkk/labs/internal/user/transport/api"

	"alexdenkk/labs/pkg/config"
	"alexdenkk/labs/pkg/middleware"

	"alexdenkk/labs/pkg/token/jwt"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.LoadEnv,
			mux.NewRouter,

			NewJwtConfig,
			NewAgentConfig,
			NewDBConfig,

			Connect,
			middleware.NewMiddleware,
			jwt.NewTokenManager,

			user_repository.New,
			user_app.New,
			user_api.New,

			lab_repository.New,
			lab_agent.New,
			lab_app.New,
			lab_api.New,
		),

		fx.Invoke(
			AutoMigrate,
			RegisterEndpoints,
			StartServer,
		),
	)

	app.Run()
}

func NewJwtConfig(cfg *config.Config) *config.JwtConfig {
	return cfg.JwtConfig
}

func NewAgentConfig(cfg *config.Config) *config.AgentConfig {
	return cfg.AgentConfig
}

func NewDBConfig(cfg *config.Config) *config.DBConfig {
	return cfg.DBConfig
}

func Connect(cfg *config.DBConfig) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil
	}

	if err := sqlDB.Ping(); err != nil {
		return nil
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&user_domain.User{})
	db.AutoMigrate(&lab_domain.Lab{})
}

func RegisterEndpoints(
	router *mux.Router,
	cfg *config.Config,
	userAPI *user_api.API,
	labAPI *lab_api.API,
) {
	router.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("./files"))))

	userAPI.RegisterEndpoints(router)
	labAPI.RegisterEndpoints(router)
}

// Middleware для CORS запросов
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Запуск HTTP сервера
func StartServer(
	lc fx.Lifecycle,
	router *mux.Router,
	cfg *config.Config,
) {
	server := &http.Server{
		Addr:         cfg.HttpConfig.Host,
		Handler:      corsMiddleware(router),
		ReadTimeout:  cfg.HttpConfig.ReadTimeout,
		WriteTimeout: cfg.HttpConfig.WriteTimeout,
		IdleTimeout:  cfg.HttpConfig.IdleTimeout,
	}

	router.HandleFunc("/health/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "i love brainfuck!",
		})
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalln(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return server.Shutdown(shutdownCtx)
		},
	})
}
