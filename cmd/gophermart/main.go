package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/OlegVankov/fantastic-engine/internal/accrual"
	"github.com/OlegVankov/fantastic-engine/internal/handler"
)

func main() {

	var (
		serverAddr  string
		accrualAddr string
	)

	flag.StringVar(&serverAddr, "a", "localhost:8080", "адрес и порт запуска сервиса")
	flag.StringVar(&accrualAddr, "r", "localhost:34567", "адрес системы расчёта начислений")

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		serverAddr = envRunAddr
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		accrualAddr = envAccrualAddr
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

	go accrual.SendAccrual(accrualAddr)

	fmt.Println("start server:", serverAddr)
	http.ListenAndServe(serverAddr, router)
}
