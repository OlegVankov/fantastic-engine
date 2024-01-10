package main

import (
	"flag"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/OlegVankov/fantastic-engine/internal"
)

var (
	serverAddr  string
	accrualAddr string
	databaseURI string
)

func main() {

	flag.StringVar(&serverAddr, "a", "localhost:8080", "адрес и порт запуска сервиса")
	flag.StringVar(&accrualAddr, "r", "http://localhost:34567", "адрес системы расчёта начислений")
	flag.StringVar(&databaseURI, "d", "", "адрес подключения к базе данных")

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		serverAddr = envRunAddr
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		accrualAddr = envAccrualAddr
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		databaseURI = envDatabaseURI
	}

	server := internal.NewServer(serverAddr, databaseURI)
	server.Run(accrualAddr)
}
