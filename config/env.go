package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	URI           string
	DB            string        // nombre de la base de datos
	RateLimit     float64       // tokens por segundo
	RateBurst     int           // capacidad máxima del balde
	LogDir        string        // directorio para archivos de log
	Env           string        // entorno de ejecución
	JWTSecret     string        // firma del token
	JWTExpiration time.Duration // cuánto dura el token
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:          getEnv("PORT", "8080"),
		URI:           RequiredEnv("MONGO_URI"),
		DB:            getEnv("MONGO_DB", "go_meli"),
		RateLimit:     getEnvFloat("RATE_LIMIT", 5),
		RateBurst:     getEnvInt("RATE_BURST", 10),
		LogDir:        getEnv("LOG_DIR", "logs"),
		Env:           getEnv("ENV", "development"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTExpiration: getEnvDuration("JWT_EXPIRATION", 1*time.Hour),
	}
}

func getEnv(key, defaultValue string) string {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	return env
}

func getEnvFloat(key string, defaultValue float64) float64 {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	val, err := strconv.ParseFloat(env, 64)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvInt(key string, defaultValue int) int {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(env)
	if err != nil {
		return defaultValue
	}
	return val
}

func RequiredEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic("environment variable " + key + " not set")
	}
	return value
}

func (c *Config) IsProd() bool {
	return c.Env == "production"
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	val, err := time.ParseDuration(env)
	if err != nil {
		return defaultValue
	}
	return val
}
