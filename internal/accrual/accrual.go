package accrual

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/OlegVankov/fantastic-engine/internal/handler"
)

func SendAccrual(addr string) {
	client := resty.New()
	url := addr + "/api/orders/"
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
}
