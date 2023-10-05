package queries

import (
	"context"
	"github.com/jackc/pgx/v4"
	"time"
)

func SelectUserBalance(ctx context.Context, conn *pgx.Conn, username string) (float64, error) {
	var balance float64
	err := conn.QueryRow(ctx, "SELECT points FROM loyalty_balance WHERE user_id = (SELECT user_id FROM users WHERE login = $1)", username).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func SelectTotalWithdrawn(ctx context.Context, conn *pgx.Conn, username string) (*float64, error) {
	var totalWithdrawn *float64
	err := conn.QueryRow(ctx, "SELECT SUM(withdrawn_points) FROM loyalty_withdrawals WHERE user_id = (SELECT user_id FROM users WHERE login = $1)", username).Scan(&totalWithdrawn)
	if err != nil {
		return nil, err
	}
	return totalWithdrawn, nil
}

func CheckOrderExists(ctx context.Context, conn *pgx.Conn, orderNumber string) (bool, error) {
	var orderExists bool
	err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM orders WHERE order_number = $1)", orderNumber).Scan(&orderExists)
	if err != nil {
		return false, err
	}
	return orderExists, nil
}

func UpdateOrder(ctx context.Context, conn *pgx.Conn, orderNumber string, amount float64) error {
	_, err := conn.Exec(ctx, "UPDATE orders SET accrual = $1, timestamp = $2, order_status = 'PROCESSING' WHERE order_number = $3", amount, time.Now(), orderNumber)
	return err
}

func InsertNewOrder(ctx context.Context, conn *pgx.Conn, orderNumber, username string) error {
	_, err := conn.Exec(ctx, "INSERT INTO orders (order_number, user_id, order_status, timestamp) VALUES ($1, (SELECT user_id FROM users WHERE login = $2), $3, $4)", orderNumber, username, "NEW", time.Now())
	return err
}

func InsertWithdrawalData(ctx context.Context, conn *pgx.Conn, username, orderNumber string, amount float64) error {
	_, err := conn.Exec(ctx, "INSERT INTO loyalty_withdrawals (user_id, order_id, withdrawn_points) VALUES ((SELECT user_id FROM users WHERE login = $1), (SELECT order_id FROM orders WHERE order_number = $2), $3)", username, orderNumber, amount)
	return err
}

func UpdateUserBalance(ctx context.Context, conn *pgx.Conn, username string, amount float64) error {
	_, err := conn.Exec(ctx, "UPDATE loyalty_balance SET points = points - $1 WHERE user_id = (SELECT user_id FROM users WHERE login = $2)", amount, username)
	return err
}
