package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	LiveKit  LiveKitConfig
	AI       AIConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	PostgresURL string
	MongoURL    string
	MongoDB     string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	ExpiryHour int
}

type LiveKitConfig struct {
	URL       string
	APIKey    string
	APISecret string
}

type AIConfig struct {
	TranslationURL string
	CaptionURL     string
	SummaryURL     string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			PostgresURL: getEnv("POSTGRES_URL", "postgres://shalom:secret@localhost:5432/shalom?sslmode=disable"),
			MongoURL:    getEnv("MONGO_URL", "mongodb://localhost:27017"),
			MongoDB:     getEnv("MONGO_DB", "shalom_chat"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "shalom-secret-key-change-in-production"),
			ExpiryHour: getEnvInt("JWT_EXPIRY_HOURS", 72),
		},
		LiveKit: LiveKitConfig{
			URL:       getEnv("LIVEKIT_URL", "ws://localhost:7880"),
			APIKey:    getEnv("LIVEKIT_API_KEY", "devkey"),
			APISecret: getEnv("LIVEKIT_API_SECRET", "devsecret"),
		},
		AI: AIConfig{
			TranslationURL: getEnv("AI_TRANSLATION_URL", "http://localhost:5001"),
			CaptionURL:     getEnv("AI_CAPTION_URL", "http://localhost:5002"),
			SummaryURL:     getEnv("AI_SUMMARY_URL", "http://localhost:5003"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
