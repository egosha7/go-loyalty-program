package handlers

import (
	"encoding/json"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func BalanceHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, logger *zap.Logger) {
	// Логируем начало обработки запроса
	logger.Info("Handling BalanceHandler")

	// Извлечение имени пользователя из куки
	cookie, err := r.Cookie("auth")
	if err != nil {
		logger.Error("Пользователь не авторизован", zap.Error(err))
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}
	username := cookie.Value

	// Запрос баланса пользователя
	var currentBalance int
	var totalWithdrawn int

	// Запрос текущего баланса из таблицы loyalty_balance
	err = conn.QueryRow(r.Context(), "SELECT points FROM loyalty_balance WHERE user_id = (SELECT user_id FROM users WHERE login = $1)", username).Scan(&currentBalance)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	// Запрос суммы снятых средств (withdrawn) из таблицы loyalty_withdrawals
	err = conn.QueryRow(r.Context(), "SELECT SUM(withdrawn_points) FROM loyalty_withdrawals WHERE user_id = (SELECT user_id FROM users WHERE login = $1)", username).Scan(&totalWithdrawn)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	// Формирование ответа
	balanceData := struct {
		Current   int `json:"current"`
		Withdrawn int `json:"withdrawn"`
	}{
		Current:   currentBalance,
		Withdrawn: totalWithdrawn,
	}

	// Отправка ответа клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(balanceData); err != nil {
		logger.Error("Ошибка при кодировании ответа", zap.Error(err))
		http.Error(w, "Ошибка при кодировании ответа", http.StatusInternalServerError)
		return
	}

	// Логируем успешное завершение обработки запроса
	logger.Info("BalanceHandler completed")
}

func WithdrawHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, logger *zap.Logger) {
	// Логируем начало обработки запроса
	logger.Info("Handling WithdrawHandler")

	// Извлечение имени пользователя из куки
	cookie, err := r.Cookie("auth")
	if err != nil {
		logger.Error("Пользователь не авторизован", zap.Error(err))
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}
	username := cookie.Value

	// Парсинг JSON-запроса
	var withdrawRequest struct {
		Order string `json:"order"`
		Sum   int    `json:"sum"`
	}

	err = json.NewDecoder(r.Body).Decode(&withdrawRequest)
	if err != nil {
		logger.Error("Неверный формат запроса", zap.Error(err))
		http.Error(w, "Неверный формат запроса", http.StatusUnprocessableEntity)
		return
	}

	// Проверка, принадлежит ли заказ пользователю
	var orderUser string
	err = conn.QueryRow(r.Context(), "SELECT users.login FROM orders\nINNER JOIN users ON users.user_id = orders.user_id\nWHERE order_number = $1", withdrawRequest.Order).Scan(&orderUser)
	if err != nil {
		logger.Error("Заказ не найден", zap.Error(err))
		http.Error(w, "Заказ не найден", http.StatusUnprocessableEntity)
		return
	}

	if orderUser != username {
		logger.Error("Заказ не принадлежит пользователю")
		http.Error(w, "Заказ не принадлежит пользователю", http.StatusForbidden)
		return
	}

	// Проверка на наличие достаточного баланса для списания
	var currentBalance int
	err = conn.QueryRow(r.Context(), "SELECT points FROM loyalty_balance WHERE user_id = (SELECT user_id FROM users WHERE login = $1)", username).Scan(&currentBalance)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	if currentBalance < withdrawRequest.Sum {
		logger.Error("На счету недостаточно средств")
		http.Error(w, "На счету недостаточно средств", http.StatusPaymentRequired)
		return
	}

	// Вставка данных о списании в таблицу loyalty_withdrawals
	_, err = conn.Exec(r.Context(), "INSERT INTO loyalty_withdrawals (user_id, order_id, withdrawn_points) VALUES ((SELECT user_id FROM users WHERE login = $1), (SELECT order_id FROM orders WHERE order_number = $2), $3)", username, withdrawRequest.Order, withdrawRequest.Sum)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	// Обновление баланса пользователя
	_, err = conn.Exec(r.Context(), "UPDATE loyalty_balance SET points = points - $1 WHERE user_id = (SELECT user_id FROM users WHERE login = $2)", withdrawRequest.Sum, username)
	if err != nil {
		logger.Error("Ошибка при обновлении баланса пользователя", zap.Error(err))
		http.Error(w, "Ошибка при обновлении баланса пользователя", http.StatusInternalServerError)
		return
	}

	// Отправка успешного ответа
	w.WriteHeader(http.StatusOK)

	// Логируем успешное завершение обработки запроса
	logger.Info("WithdrawHandler completed")
}

func WithdrawalsHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, logger *zap.Logger) {
	// Логируем начало обработки запроса
	logger.Info("Handling WithdrawalsHandler")

	// Извлечение имени пользователя из куки
	cookie, err := r.Cookie("auth")
	if err != nil {
		logger.Error("Пользователь не авторизован", zap.Error(err))
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}
	username := cookie.Value

	// Запрос данных о списаниях пользователя
	rows, err := conn.Query(r.Context(), "SELECT orders.order_number, loyalty_withdrawals.withdrawn_points, orders.timestamp FROM orders INNER JOIN loyalty_withdrawals ON orders.order_id = loyalty_withdrawals.order_id WHERE orders.user_id = (SELECT user_id FROM users WHERE login = $1) ORDER BY orders.timestamp ASC", username)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создание списка для хранения данных о списаниях
	var withdrawals []struct {
		Order       string `json:"order"`
		Sum         int    `json:"sum"`
		ProcessedAt string `json:"processed_at"`
	}

	// Итерация по результатам запроса и добавление их в список withdrawals
	for rows.Next() {
		var orderID string
		var withdrawnPoints int
		var withdrawnTime time.Time

		err := rows.Scan(&orderID, &withdrawnPoints, &withdrawnTime)
		if err != nil {
			logger.Error("Ошибка при сканировании данных", zap.Error(err))
			http.Error(w, "Ошибка при сканировании данных", http.StatusInternalServerError)
			return
		}

		// Преобразование времени к формату RFC3339
		timeStr := withdrawnTime.Format(time.RFC3339)

		withdrawal := struct {
			Order       string `json:"order"`
			Sum         int    `json:"sum"`
			ProcessedAt string `json:"processed_at"`
		}{
			Order:       orderID,
			Sum:         withdrawnPoints,
			ProcessedAt: timeStr,
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	// Проверка наличия данных о списаниях
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Отправка успешного ответа с данными о списаниях
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		logger.Error("Ошибка при кодировании ответа", zap.Error(err))
		http.Error(w, "Ошибка при кодировании ответа", http.StatusInternalServerError)
		return
	}

	// Логируем успешное завершение обработки запроса
	logger.Info("WithdrawalsHandler completed")
}
