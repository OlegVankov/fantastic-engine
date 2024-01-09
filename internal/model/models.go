package model

import "time"

type User struct {
	ID       uint64  `db:"id" json:"-"`
	Login    string  `db:"login" json:"login"`
	Password string  `db:"password" json:"-"`
	Balance  float64 `db:"balance"`
	Withdraw float64 `db:"withdraw"`
}

type Order struct {
	Number    string
	Status    string    `json:"status"`
	Accrual   float64   `json:"accrual"`
	UserLogin string    `json:"-"`
	Uploaded  time.Time `db:"uploaded" json:"uploaded_at"`
}

type Withdraw struct {
	ID          uint64    `db:"id" json:"-"`
	Number      string    `db:"number" json:"order"`
	Amount      float64   `db:"amount" json:"sum"`
	UserLogin   string    `db:"userlogin" json:"-"`
	ProcessedAt time.Time `db:"processed" json:"processed_at"`
}
