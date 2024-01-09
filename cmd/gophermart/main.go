package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/OlegVankov/fantastic-engine/internal/handler"
	"github.com/OlegVankov/fantastic-engine/internal/handler/accrual"
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

	router := chi.NewRouter()

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)

		r.Route("/", func(r chi.Router) {
			r.Use(handler.Auth)

			r.Post("/orders", handler.Orders)
			r.Get("/orders", handler.GetOrders)

			r.Post("/balance/withdraw", handler.Withdraw)
			r.Get("/balance", handler.Balance)
			r.Get("/withdrawals", handler.Withdrawals)
		})
	})

	err := handler.SetRepository(databaseURI)
	if err != nil {
		log.Fatal(err)
	}

	go accrual.SendAccrual(accrualAddr)

	http.ListenAndServe(serverAddr, router)
}
