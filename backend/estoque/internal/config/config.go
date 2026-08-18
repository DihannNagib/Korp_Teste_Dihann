package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPassword string
	DBName string
	ServerPort string
}

func Load(envFile string) *Config {
	if err := godotenv.Load(envFile); err != nil {
		log.Printf(".env não encontrado; usando variáveis de ambiente")
	}

	return &Config{
		DBHost: requiredEnv("ESTOQUE_DB_HOST"),
		DBPort: requiredEnv("ESTOQUE_DB_PORT"),
		DBUser: requiredEnv("ESTOQUE_DB_USER"),
		DBPassword: requiredEnv("ESTOQUE_DB_PASSWORD"),
		DBName: requiredEnv("ESTOQUE_DB_NAME"),
		ServerPort: requiredEnv("ESTOQUE_API_PORT"),
	}
}

func requiredEnv(key string) string {
	value, exists := os.LookupEnv(key)

	if !exists || value == "" {
		log.Fatalf("variável de ambiente obrigatória não definida: %s", key)
	}

	return value
}