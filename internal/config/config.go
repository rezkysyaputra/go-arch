package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Log      LogConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	Redis    RedisConfig
}

type AppConfig struct {
	Env string
}

type LogConfig struct {
	Level  string
	Pretty bool
}

type HTTPConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

func (c DatabaseConfig) MigrationURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

type RabbitMQConfig struct {
	URL                   string
	Exchange              string
	UserCreatedQueue      string
	UserCreatedRoutingKey string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func Load() (Config, error) {
	setDefaults()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if !isMissingConfigFile(err) {
			return Config{}, err
		}
	}

	return Config{
		App: AppConfig{
			Env: viper.GetString("APP_ENV"),
		},
		Log: LogConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Pretty: viper.GetBool("LOG_PRETTY"),
		},
		HTTP: HTTPConfig{
			Port: viper.GetString("HTTP_PORT"),
		},
		Database: DatabaseConfig{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetString("DB_PORT"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			Name:            viper.GetString("DB_NAME"),
			SSLMode:         viper.GetString("DB_SSLMODE"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:                   viper.GetString("RABBITMQ_URL"),
			Exchange:              viper.GetString("RABBITMQ_EXCHANGE"),
			UserCreatedQueue:      viper.GetString("RABBITMQ_USER_CREATED_QUEUE"),
			UserCreatedRoutingKey: viper.GetString("RABBITMQ_USER_CREATED_ROUTING_KEY"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
	}, nil
}

func setDefaults() {
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_PRETTY", true)
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "go_arch")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 30*time.Minute)
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("RABBITMQ_EXCHANGE", "go_arch.events")
	viper.SetDefault("RABBITMQ_USER_CREATED_QUEUE", "user.created.queue")
	viper.SetDefault("RABBITMQ_USER_CREATED_ROUTING_KEY", "user.created")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
}

func isMissingConfigFile(err error) bool {
	if errors.Is(err, os.ErrNotExist) {
		return true
	}

	_, ok := err.(viper.ConfigFileNotFoundError)
	return ok
}
