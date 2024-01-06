package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"

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

	go func() {
		client := resty.New()
		url := accrualAddr + "/api/orders/"
		ball := struct {
			Order   string  `json:"order"`
			Status  string  `json:"status"`
			Accrual float64 `json:"accrual"`
		}{}
		for {
			for k := range handler.Orders2 {
				resp, err := client.R().SetResult(&ball).Get(url + k)
				if err != nil {
					fmt.Printf("[ERROR] %s\n", err.Error())
				}

				if resp.StatusCode() == http.StatusOK {

					username := handler.Orders2[ball.Order]
					user := handler.Users2[username]
					order := handler.Users2[username].Order[ball.Order]
					order.Status = ball.Status

					if ball.Status == "PROCESSED" {
						order.Accrual = ball.Accrual
						user.Balance += ball.Accrual
					}

					handler.Users2[username] = user
					handler.Users2[username].Order[ball.Order] = order

				}

			}

			<-time.After(time.Second)
		}
	}()

	fmt.Println("start server:", serverAddr)
	http.ListenAndServe(serverAddr, router)
}
