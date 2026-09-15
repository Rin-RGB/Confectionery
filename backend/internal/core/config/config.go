package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	LoggerCfg *LoggerConfig
	ServerCfg *ServerConfig
	DbConfig  *DatabaseConfig
	AuthCfg   *AuthConfig
}

type LoggerConfig struct {
	Level  string `envconfig:"LOGGER_LEVEL" default:"DEBUG"`
	Folder string `envconfig:"LOGGER_FOLDER"`
}

type ServerConfig struct {
	Addr              string        `envconfig:"SERVER_ADDR" default:":8080"`
	ReadHeaderTimeout time.Duration `envconfig:"SERVER_READ_HEADER_TIMEOUT" default:"5s"`
	ShutdownTimeout   time.Duration `envconfig:"SERVER_SHUTDOWN_TIMEOUT" default:"5s"`
}

type DatabaseConfig struct {
	User     string        `envconfig:"POSTGRES_USER" default:"postgres"`
	Password string        `envconfig:"POSTGRES_PASSWORD" default:"password"`
	Name     string        `envconfig:"POSTGRES_DB" default:"postgres"`
	Host     string        `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     string        `envconfig:"POSTGRES_PORT" default:"5432"`
	Timeout  time.Duration `envconfig:"DATABASE_TIMEOUT" default:"5s"`
	SSLMode  string        `envconfig:"DATABASE_SSL_MODE" default:"disable"`
}

type AuthConfig struct {
	SigningKey      string        `envconfig:"JWT_SIGNING_KEY" default:"development-secret"`
	AccessTokenTTL  time.Duration `envconfig:"JWT_ACCESS_TOKEN_TTL" default:"15m"`
	RefreshTokenTTL time.Duration `envconfig:"JWT_REFRESH_TOKEN_TTL" default:"720h"`
	BcryptCost      int           `envconfig:"BCRYPT_COST" default:"12"`
}

func NewConfig() (*Config, error) {
	config := Config{}
	if err := envconfig.Process("", &config); err != nil {
		return &Config{}, err
	}

	return &config, nil
}

func NewConfigMust() *Config {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}

	return config
}
