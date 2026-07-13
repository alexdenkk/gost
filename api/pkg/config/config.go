package config

import (
	"os"
	"time"
)

type Config struct {
	AgentConfig *AgentConfig
	JwtConfig   *JwtConfig
	DBConfig    *DBConfig
	HttpConfig  *HttpConfig
}

type JwtConfig struct {
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type AgentConfig struct {
	AccessToken       string
	BaseURL           string
	GeneratorAccessID string
	FormatterAccessID string
}

type HttpConfig struct {
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func LoadEnv() *Config {
	return &Config{
		AgentConfig: &AgentConfig{
			AccessToken:       os.Getenv("AGENT_ACCESS_TOKEN"),
			BaseURL:           os.Getenv("AGENT_BASE_URL"),
			GeneratorAccessID: os.Getenv("AGENT_GENERATOR_ACCESS_ID"),
			FormatterAccessID: os.Getenv("AGENT_FORMATTER_ACCESS_ID"),
		},

		DBConfig: &DBConfig{
			Name:     os.Getenv("DB_NAME"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},

		JwtConfig: &JwtConfig{
			AccessSecret:  []byte(os.Getenv("JWT_ACCESS_SECRET")),
			RefreshSecret: []byte(os.Getenv("JWT_REFRESH_SECRET")),
			AccessTTL:     15 * time.Minute,
			RefreshTTL:    20 * time.Minute,
		},

		HttpConfig: &HttpConfig{
			Host:         os.Getenv("HOST") + ":" + os.Getenv("PORT"),
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}
