package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/OlegVankov/fantastic-engine/internal/repository"
	"github.com/OlegVankov/fantastic-engine/internal/util/hash"
	"github.com/OlegVankov/fantastic-engine/internal/util/jwt"
	"github.com/OlegVankov/fantastic-engine/internal/util/lun"
)

type Handler struct {
	Repository repository.Repository
}

type credential struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {

	c := credential{}

	err := json.NewDecoder(r.Body).Decode(&c)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pass, err := hash.StringToHash(c.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := h.Repository.AddUser(r.Context(), c.Login, pass)
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

	tkn, err := jwt.CreateToken(user.Login, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	authorization := fmt.Sprintf("Bearer %s", tkn)

	w.Header().Add("Authorization", authorization)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	c := credential{}

	err := json.NewDecoder(r.Body).Decode(&c)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.Repository.GetUser(r.Context(), c.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !hash.CheckPassword(user.Password, c.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tkn, err := jwt.CreateToken(user.Login, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	authorization := fmt.Sprintf("Bearer %s", tkn)

	w.Header().Add("Authorization", authorization)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
	}

	number := string(body)

	if !lun.CheckLun(number) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	username := r.Header.Get("username")

	_, err = h.Repository.AddOrder(r.Context(), username, number)
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == "23505" {
			order, err := h.Repository.GetOrderByNumber(r.Context(), number)
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

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")

	orders, err := h.Repository.GetOrdersByLogin(r.Context(), username)
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

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
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

	if !lun.CheckLun(withdraw.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = h.Repository.UpdateWithdraw(r.Context(), username, withdraw.Order, withdraw.Sum)
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

func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")

	user, err := h.Repository.GetBalance(r.Context(), username)
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

func (h *Handler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")

	wd, err := h.Repository.GetWithdrawals(r.Context(), username)
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
