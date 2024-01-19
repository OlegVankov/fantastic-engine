package accrual

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/OlegVankov/fantastic-engine/internal/handler"
)

func SendAccrual(addr string, handler *handler.Handler) {
	ctx := context.Background()
	client := resty.New()
	url := addr + "/api/orders/"
	ball := struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual"`
	}{}
	for {

		orders, err := handler.Repository.GetOrders(ctx)

		if err != nil {
			continue
		}

		for _, k := range orders {
			url := url + k.Number
			resp, err := client.R().
				SetResult(&ball).
				Get(url)
			if err != nil {
				fmt.Printf("[ERROR] %s\n", err.Error())
			}

			if resp.StatusCode() == http.StatusOK {

				err := handler.Repository.UpdateOrder(ctx, ball.Order, ball.Status, ball.Accrual)
				if err != nil {
					fmt.Printf("[ERROR] %s\n", err.Error())
					continue
				}

			}

		}

		<-time.After(time.Second * time.Duration(5))
	}
}
