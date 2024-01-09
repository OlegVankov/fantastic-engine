package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/OlegVankov/fantastic-engine/internal/repository"
	"github.com/OlegVankov/fantastic-engine/internal/repository/postgres"
	"github.com/OlegVankov/fantastic-engine/internal/util"
)

type credential struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var (
	Repository repository.Repository
	// = postgres.NewUserRepository("postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable")
)

func SetRepository(dsn string) error {
	Repository = postgres.NewUserRepository(dsn)
	return nil
}

func Register(w http.ResponseWriter, r *http.Request) {

	c := credential{}

	err := json.NewDecoder(r.Body).Decode(&c)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := Repository.AddUser(context.Background(), c.Login, c.Password)
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == "23505" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		fmt.Printf("%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tkn, err := util.CreateToken(user.Login, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	authorization := fmt.Sprintf("Bearer %s", tkn)

	w.Header().Add("Authorization", authorization)
	w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	c := credential{}

	err := json.NewDecoder(r.Body).Decode(&c)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := Repository.GetUser(context.Background(), c.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.Password != c.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tkn, _ := util.CreateToken(user.Login, user.ID)
	authorization := fmt.Sprintf("Bearer %s", tkn)

	w.Header().Add("Authorization", authorization)
	w.WriteHeader(http.StatusOK)
}

func Orders(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
	}

	number := string(body)

	if !util.CheckLun(number) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	username := r.Header.Get("username")

	_, err = Repository.AddOrder(context.Background(), username, number)
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == "23505" {
			order, err := Repository.GetOrderByNumber(context.Background(), number)
			// fmt.Printf("username: %s number %s %v\n", username, number, order)
			if err == nil {
				if order.UserLogin == username {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")

	orders, err := Repository.GetOrdersByLogin(context.Background(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")
	withdraw := struct {
		Order string
		Sum   float64
	}{}
	err := json.NewDecoder(r.Body).Decode(&withdraw)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !util.CheckLun(withdraw.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = Repository.UpdateWithdraw(context.Background(), username, withdraw.Order, withdraw.Sum)
	if err != nil {
		if err.Error() == "balance error" {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func Balance(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")

	user, err := Repository.GetBalance(context.Background(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	balance := struct {
		Current   float64
		Withdrawn float64
	}{
		user.Balance,
		user.Withdraw,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func Withdrawals(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")

	wd, err := Repository.GetWithdrawals(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(wd) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wd)
}
