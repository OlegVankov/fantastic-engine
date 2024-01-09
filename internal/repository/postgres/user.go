package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/OlegVankov/fantastic-engine/internal/model"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(dsn string) *UserRepository {
	db, _ := sqlx.Open("pgx", dsn)
	repo := &UserRepository{
		db: db,
	}
	err := repo.Bootstrap(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return repo
}

func (r *UserRepository) Bootstrap(ctx context.Context) error {
	users := `CREATE TABLE IF NOT EXISTS users
(
        id bigserial PRIMARY KEY,
        login varchar not null,
        password varchar not null,
        balance decimal(10, 2) default 0,
        withdraw decimal(10, 2) default 0,
        constraint users_unique_login unique (login)
);
`
	orders := `CREATE TABLE IF NOT EXISTS orders
(
    number varchar PRIMARY KEY not null,
    status varchar,
	accrual decimal(10, 2) default 0,
	userlogin varchar not null,
	uploaded timestamp with time zone not null default now(),
    constraint orders_unique_number unique (number)
);
`
	withdraw := `CREATE TABLE IF NOT EXISTS withdraw
(
    id bigserial PRIMARY KEY,
    number varchar not null,
    amount decimal(10, 2) default 0,
    userlogin varchar not null,
    processed timestamp with time zone not null default now()
);
`
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = r.db.ExecContext(ctx, users)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, orders)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, withdraw)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) AddUser(ctx context.Context, login, password string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(ctx,
		"insert into users(login, password) values($1, $2) returning *;",
		login,
		password,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Balance,
		&user.Withdraw,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUser(ctx context.Context, login string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(ctx,
		"select * from users where login = $1",
		login,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Balance,
		&user.Withdraw,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) AddOrder(ctx context.Context, login, number string) (*model.Order, error) {
	order := &model.Order{}
	err := r.db.QueryRowContext(ctx,
		"insert into orders(number, userlogin, status) values($1, $2, $3) returning *;",
		number,
		login,
		"NEW",
	).Scan(
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UserLogin,
		&order.Uploaded,
	)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *UserRepository) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	order := &model.Order{}
	err := r.db.QueryRowContext(ctx,
		"select number, userlogin from orders where number = $1",
		number,
	).Scan(
		&order.Number,
		&order.UserLogin,
	)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *UserRepository) UpdateOrder(ctx context.Context, number, status string, accrual float64) error {
	order := &model.Order{}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow("update orders set status = $1, accrual = $2 where number = $3 returning number, userlogin",
		status,
		accrual,
		number,
	).Scan(
		&order.Number,
		&order.UserLogin,
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec("update users set balance = balance + $1 where login = $2",
		accrual,
		order.UserLogin,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) GetOrdersByLogin(ctx context.Context, username string) ([]model.Order, error) {
	orders := []model.Order{}
	err := r.db.SelectContext(ctx, &orders, "select * from orders where userlogin = $1", username)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *UserRepository) GetBalance(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(ctx,
		"select balance, withdraw from users where login = $1",
		username,
	).Scan(
		&user.Balance,
		&user.Withdraw,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetOrders(ctx context.Context) ([]model.Order, error) {
	orders := []model.Order{}
	err := r.db.SelectContext(ctx, &orders, "select * from orders")
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *UserRepository) UpdateWithdraw(ctx context.Context, login, number string, sum float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	order := &model.Order{}
	balance := 0.0
	tx.QueryRow("select balance from users where login = $1", login).Scan(&balance)

	if balance < sum {
		return errors.New("balance error")
	}

	_, err = tx.Exec("update users set balance = balance - $1, withdraw = withdraw + $1 where login = $2",
		sum,
		login,
	)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx,
		"insert into orders(number, userlogin, status) values($1, $2, $3) returning *;",
		number,
		login,
		"NEW",
	).Scan(
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UserLogin,
		&order.Uploaded,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec("insert into withdraw(number, amount, userlogin) values($1, $2, $3);",
		number,
		sum,
		login,
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) GetWithdrawals(ctx context.Context, login string) ([]model.Withdraw, error) {
	withdrawals := []model.Withdraw{}
	err := r.db.SelectContext(ctx, &withdrawals, "select * from withdraw where userlogin = $1 order by processed;", login)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}
