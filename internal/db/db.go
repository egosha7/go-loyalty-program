package db

import (
	"context"
	"fmt"
	"github.com/egosha7/go-loyalty-program.git/internal/config"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"net/http"
)

func ConnectToDB(cfg *config.Config) (*pgx.Conn, error) {

	if cfg.DataBaseURI == "" {
		// Возвращаем nil, если строка подключения пуста
		conn := &pgx.Conn{}
		return conn, nil
	}

	connConfig, err := pgx.ParseConfig(cfg.DataBaseURI)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func PingDB(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	err := conn.Ping(context.Background())
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Check(logger *zap.Logger, conn *pgx.Conn) {
	// Проверка существования таблиц и создание их при необходимости
	tables := []struct {
		name      string
		createSQL string
	}{
		{"users", `
			CREATE TABLE IF NOT EXISTS users (
				user_id serial PRIMARY KEY,
				login text,
				password text
			);
		`},
		{"orders", `
			CREATE TABLE IF NOT EXISTS orders (
				order_id serial PRIMARY KEY,
				user_id integer REFERENCES users (user_id),
				order_status text,
				order_number text,
				"timestamp" timestamptz
			);
		`},
		{"loyalty_withdrawals", `
			CREATE TABLE IF NOT EXISTS loyalty_withdrawals (
				withdrawal_id serial PRIMARY KEY,
				user_id integer REFERENCES users (user_id),
				order_id integer REFERENCES orders (order_id),
				withdrawn_points double precision
			);
		`},

		{"loyalty_balance", `
			CREATE TABLE IF NOT EXISTS loyalty_balance (
				loyalty_id serial PRIMARY KEY,
				user_id integer REFERENCES users (user_id),
				points double precision
			);
		`},
	}

	for _, table := range tables {
		err := createTableIfNotExists(conn, table.createSQL)
		if err != nil {
			logger.Error("Ошибка запроса на создание таблицы "+table.name, zap.Error(err))
		} else {
			logger.Info("Создана таблица " + table.name)
		}
	}

	fmt.Println("Таблицы успешно созданы или уже существуют.")
}

// createTableIfNotExists создает таблицу, если она не существует
func createTableIfNotExists(conn *pgx.Conn, createSQL string) error {
	_, err := conn.Exec(context.Background(), createSQL)
	return err
}
