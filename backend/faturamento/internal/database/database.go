package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/config"
)

func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("erro ao obter conexão SQL: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("erro ao verificar conexão com banco: %v", err)
	}

	log.Println("conectado ao banco de dados com sucesso")

	return db
}