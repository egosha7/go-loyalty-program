package queries

import (
	"context"
	"github.com/jackc/pgx/v4"
)

func CheckUniqueLogin(ctx context.Context, conn *pgx.Conn, login string) (bool, error) {
	var existingUser string
	err := conn.QueryRow(ctx, "SELECT login FROM users WHERE login = $1", login).Scan(&existingUser)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func InsertUser(ctx context.Context, conn *pgx.Conn, login string, hashedPassword []byte) (int, error) {
	var userID int
	err := conn.QueryRow(ctx, "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id", login, hashedPassword).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func InsertLoyaltyBalance(ctx context.Context, conn *pgx.Conn, userID int) error {
	_, err := conn.Exec(ctx, "INSERT INTO loyalty_balance (user_id, points) VALUES ($1, $2)", userID, 0.0)
	if err != nil {
		return err
	}
	return nil
}

func LoginUser(ctx context.Context, conn *pgx.Conn, login string) (string, error) {
	var hashedPassword string
	err := conn.QueryRow(ctx, "SELECT password FROM users WHERE login = $1", login).Scan(&hashedPassword)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}
