package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

type UserClaim struct {
	jwt.RegisteredClaims
	Username string
}

// [login]
var Users2 = map[string]User{}

// [login]
var Orders2 = map[string]string{}

func createToken(username string) (string, error) {
	userClaim := &UserClaim{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaim)

	tokenString, err := token.SignedString([]byte("secret_key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func checkLun(num string) bool {
	sum := 0
	parity := len(num) % 2

	for i, v := range num {
		digit, _ := strconv.Atoi(string(v))
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}
		sum += digit
	}

	return sum%10 == 0
}

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

	tkn, _ := createToken(user.Login)
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

	tkn, _ := createToken(user.Login)
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
	fmt.Printf("[INFO] method: %s path: %s body: %q\n", r.Method, r.URL.Path, body)

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
	}

	number := string(body)

	if !checkLun(number) {
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

	// user := Users2[username]
	// user.Order = append(user.Order, Order{Number: number, Status: "NEW", Uploaded: time.Now()})
	// Users2[username] = user

	Users2[username].Order[number] = Order{Number: number, Status: "NEW", Uploaded: time.Now()}

	// fmt.Printf("[INFO] POST /api/user/Orders2 %s %s %v\n", username, number, Users2[username].Order)

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

	if !checkLun(withdraw.Order) {
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
		withdrawals = append(withdrawals, Wd{Order: v.Number, Sum: Users2[username].Withdraw, Proccessed_At: time.Now(), uploaded: v.Uploaded})
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

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userClaim := UserClaim{}

		token, err := jwt.ParseWithClaims(auth, &userClaim, func(token *jwt.Token) (interface{}, error) { return []byte("secret_key"), nil })

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if _, ok := Users2[userClaim.Username]; !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Add("username", userClaim.Username)
		h.ServeHTTP(w, r)
	})
}
