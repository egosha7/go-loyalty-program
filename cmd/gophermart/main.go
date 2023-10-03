package main

import (
	"context"
	"fmt"
	"github.com/egosha7/go-loyalty-program.git/internal/config"
	"github.com/egosha7/go-loyalty-program.git/internal/db"
	"github.com/egosha7/go-loyalty-program.git/internal/loger"
	"github.com/egosha7/go-loyalty-program.git/internal/routes"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Проверка конфигурации флагов и переменных окружения
	cfg := config.OnFlag(logger)

	conn, err := db.ConnectToDB(cfg)
	if err != nil {
		logger.Error("Error connecting to database", zap.Error(err))
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	exists, err := db.CheckDatabaseExists(conn, logger)
	if err != nil {
		logger.Fatal("Ошибка при проверке наличия базы данных", zap.Error(err))
	}

	if !exists {
		// Если базы данных нет, восстановить ее из резервной копии
		err := db.RestoreDatabase()
		if err != nil {
			logger.Fatal("Ошибка при восстановлении базы данных из резервной копии", zap.Error(err))
		}
		logger.Info("База данных восстановлена из резервной копии")
	}

	r := routes.SetupRoutes(cfg, conn, logger)

	// Запуск сервера
	err = http.ListenAndServe(cfg.Addr, loger.LogMiddleware(logger, r))
	if err != nil {
		logger.Error("Error starting server", zap.Error(err))
		os.Exit(1)
	}
}
