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

func Load() *Config {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf(".env não encontrado; usando variáveis de ambiente")
	}

	return &Config{
		DBHost: requiredEnv("FATURAMENTO_DB_HOST"),
		DBPort: requiredEnv("FATURAMENTO_DB_PORT"),
		DBUser: requiredEnv("FATURAMENTO_DB_USER"),
		DBPassword: requiredEnv("FATURAMENTO_DB_PASSWORD"),
		DBName: requiredEnv("FATURAMENTO_DB_NAME"),
		ServerPort: requiredEnv("FATURAMENTO_API_PORT"),
	}
}

func requiredEnv(key string) string {
	value, exists := os.LookupEnv(key)

	if !exists || value == "" {
		log.Fatalf("variável de ambiente obrigatória não definida: %s", key)
	}

	return value
}