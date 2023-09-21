package db

import (
	"context"
	"github.com/egosha7/go-loyalty-program.git/internal/config"
	"github.com/jackc/pgx/v4"
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
