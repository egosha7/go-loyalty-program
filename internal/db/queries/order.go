package queries

import (
	"context"
	"github.com/jackc/pgx/v4"
	"time"
)

func InsertOrder(ctx context.Context, conn *pgx.Conn, orderNumber string, userID int, orderStatus string, timestamp time.Time, orderAccrual float64) error {
	_, err := conn.Exec(ctx, "INSERT INTO orders (order_number, user_id, order_status, timestamp, order_accural) VALUES ($1, $2, $3, $4, $5)",
		orderNumber, userID, orderStatus, timestamp, orderAccrual)
	if err != nil {
		return err
	}
	return nil
}

func UpdateLoyaltyBalance(ctx context.Context, conn *pgx.Conn, userID int, points float64) error {
	_, err := conn.Exec(ctx, "UPDATE loyalty_balance SET points = points + $1 WHERE user_id = $2", points, userID)
	if err != nil {
		return err
	}
	return nil
}

func SelectUserIDByLogin(ctx context.Context, conn *pgx.Conn, login string) (int, error) {
	var userID int
	err := conn.QueryRow(ctx, "SELECT user_id FROM users WHERE login = $1", login).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func SelectUserIDByOrderNumber(ctx context.Context, conn *pgx.Conn, orderNumber string) (int, error) {
	var existingUser int
	err := conn.QueryRow(ctx, "SELECT user_id FROM orders WHERE order_number = $1", orderNumber).Scan(&existingUser)
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}
	return existingUser, nil
}
