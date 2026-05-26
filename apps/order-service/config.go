package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration.
type Config struct {
	AppEnv               string
	Port                 int
	DBHost               string
	DBPort               int
	DBUser               string
	DBPassword           string
	DBName               string
	KafkaBrokers         string
	KafkaSchemaRegistry  string
	RedisAddr            string
	OTelEndpoint         string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

	return &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Port:                port,
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              dbPort,
		DBUser:              getEnv("DB_USER", "order_user"),
		DBPassword:          getEnv("DB_PASSWORD", "order_pass"),
		DBName:              getEnv("DB_NAME", "order_db"),
		KafkaBrokers:        getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaSchemaRegistry: getEnv("KAFKA_SCHEMA_REGISTRY", "http://localhost:8081"),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		OTelEndpoint:        getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
