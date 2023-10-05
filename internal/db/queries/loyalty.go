package queries

import (
	"context"
	"github.com/jackc/pgx/v4"
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
