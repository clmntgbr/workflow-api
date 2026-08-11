package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL          string
	ClerkWebhookSecret   string
	Port                 string
	Environment          string
	ClerkSecretKey       string
	ClerkFrontendAPI     string
	CORSAllowedOrigins   []string
	CORSAllowCredentials bool
	CORSAllowMethods     []string
	CORSAllowHeaders     []string
	CORSMaxAge           int
	RateLimitMax         int
	RabbitMQURL          string
	RabbitMQExchange     string
	RabbitMQQueue        string
	RabbitMQRoutingKey   string
	RabbitMQRetryTTLMS   int
	WorkerMaxRetries     int
	OutboxPollInterval   time.Duration
	WorkerConcurrency    int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	return &Config{
		DatabaseURL:          getEnv("DATABASE_URL"),
		ClerkWebhookSecret:   getEnv("CLERK_WEBHOOK_SECRET"),
		Port:                 getEnv("PORT"),
		Environment:          getEnv("GO_ENV"),
		ClerkSecretKey:       getEnv("CLERK_SECRET_KEY"),
		ClerkFrontendAPI:     getEnv("CLERK_FRONTEND_API"),
		CORSAllowedOrigins:   strings.Split(getEnv("CORS_ALLOWED_ORIGINS"), ","),
		CORSAllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS"),
		CORSAllowMethods:     strings.Split(getEnv("CORS_ALLOW_METHODS"), ","),
		CORSAllowHeaders:     strings.Split(getEnv("CORS_ALLOW_HEADERS"), ","),
		CORSMaxAge:           getEnvInt("CORS_MAX_AGE"),
		RateLimitMax:         getEnvInt("RATE_LIMIT_MAX"),
		RabbitMQURL:          getEnv("RABBITMQ_URL"),
		RabbitMQExchange:     getEnvOrDefault("RABBITMQ_EXCHANGE", "domain.events"),
		RabbitMQQueue:        getEnvOrDefault("RABBITMQ_QUEUE", "user.events"),
		RabbitMQRoutingKey:   getEnvOrDefault("RABBITMQ_ROUTING_KEY", "user.#"),
		RabbitMQRetryTTLMS:   getEnvIntOrDefault("RABBITMQ_RETRY_TTL_MS", 30000),
		WorkerMaxRetries:     getEnvIntOrDefault("WORKER_MAX_RETRIES", 3),
		OutboxPollInterval:   getEnvDuration("OUTBOX_POLL_INTERVAL", 2*time.Second),
		WorkerConcurrency:    getEnvIntOrDefault("WORKER_CONCURRENCY", 4),
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	log.Panicf("required environment variable %s is not set", key)
	return ""
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string) bool {
	value := os.Getenv(key)
	if value == "" {
		return false
	}

	return value == "true"
}

func getEnvInt(key string) int {
	value := os.Getenv(key)
	if value == "" {
		log.Panicf("required environment variable %s is not set", key)
		return 0
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		log.Panicf("invalid integer for %s: %q", key, value)
		return 0
	}

	return parsedValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		log.Panicf("invalid integer for %s: %q", key, value)
		return 0
	}
	return parsedValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Panicf("invalid duration for %s: %q", key, value)
		return 0
	}
	return parsed
}
