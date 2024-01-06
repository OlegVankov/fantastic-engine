package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/OlegVankov/fantastic-engine/internal/util"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string
	Balance  float64
	Withdraw float64
	Order    map[string]Order
}

type Order struct {
	Number   string
	Status   string
	Accrual  float64
	Uploaded time.Time
}

// [login]
var Users2 = map[string]User{}

// [login]
var Orders2 = map[string]string{}

func Register(w http.ResponseWriter, r *http.Request) {

	user := User{}

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, ok := Users2[user.Login]; ok {
		w.WriteHeader(http.StatusConflict)
		return
	}

	tkn, _ := util.CreateToken(user.Login)
	authorization := fmt.Sprintf("Bearer %s", tkn)

	user.Token = tkn
	user.Order = map[string]Order{}
	Users2[user.Login] = user

	w.Header().Add("Authorization", authorization)
	w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	user := User{}

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if Users2[user.Login].Password != user.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tkn, _ := util.CreateToken(user.Login)
	authorization := fmt.Sprintf("Bearer %s", tkn)

	user.Token = tkn
	user.Order = map[string]Order{}
	Users2[user.Login] = user

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

	if _, ok := Orders2[number]; ok {
		if Orders2[number] == username {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusConflict)
		return
	}

	Orders2[number] = username

	Users2[username].Order[number] = Order{Number: number, Status: "NEW", Uploaded: time.Now()}

	w.WriteHeader(http.StatusAccepted)
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")
	o := []Order{}
	for _, order := range Users2[username].Order {
		o = append(o, order)
	}
	if len(o) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(o)
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

	user := Users2[username]

	if !util.CheckLun(withdraw.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	Orders2[withdraw.Order] = username

	if withdraw.Sum > user.Balance {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	user.Balance -= withdraw.Sum
	user.Withdraw += withdraw.Sum
	user.Order[withdraw.Order] = Order{Number: withdraw.Order, Status: "NEW", Uploaded: time.Now()}

	Users2[username] = user

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func Balance(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")

	balance := struct {
		Current   float64
		Withdrawn float64
	}{
		Current:   Users2[username].Balance,
		Withdrawn: Users2[username].Withdraw,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func Withdrawals(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")

	type Wd struct {
		Order         string
		Sum           float64
		Proccessed_At time.Time
		uploaded      time.Time
	}
	withdrawals := []Wd{}

	for _, v := range Users2[username].Order {
		withdrawals = append(withdrawals,
			Wd{
				Order:         v.Number,
				Sum:           Users2[username].Withdraw,
				Proccessed_At: time.Now(),
				uploaded:      v.Uploaded,
			})
	}

	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[j].uploaded.Before(withdrawals[i].uploaded)
	})

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}
