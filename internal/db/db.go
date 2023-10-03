package db

import (
	"context"
	"github.com/egosha7/go-loyalty-program.git/internal/config"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/exec"
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

// Путь к файлу резервной копии
const backupFilePath = "SQL.sql"
const dbName = "loyalty"

// Функция для проверки наличия базы данных
func CheckDatabaseExists(conn *pgx.Conn, logger *zap.Logger) (bool, error) {
	var exists bool
	err := conn.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		return false, err
	}
	return exists, nil
}

// Функция для восстановления базы данных из резервной копии
func RestoreDatabase() error {
	cmd := exec.Command("psql", "-U", "ваше_имя_пользователя", "-d", "ваша_база_данных", "-f", backupFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
