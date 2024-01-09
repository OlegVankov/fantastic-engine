package repository

import (
	"context"

	"github.com/OlegVankov/fantastic-engine/internal/model"
)

type Repository interface {
	AddUser(ctx context.Context, login, password string) (*model.User, error)
	GetUser(ctx context.Context, login string) (*model.User, error)
	AddOrder(ctx context.Context, login, number string) (*model.Order, error)
	GetOrdersByLogin(ctx context.Context, number string) ([]model.Order, error)
	GetOrders(ctx context.Context) ([]model.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*model.Order, error)
	UpdateOrder(ctx context.Context, number, status string, accrual float64) error
	GetBalance(ctx context.Context, username string) (*model.User, error)
	UpdateWithdraw(ctx context.Context, login, number string, sum float64) error
	GetWithdrawals(ctx context.Context, login string) ([]model.Withdraw, error)
}
