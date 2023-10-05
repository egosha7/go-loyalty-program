package handlers

import (
	"encoding/json"
	"github.com/egosha7/go-loyalty-program.git/internal/db/queries"
	"github.com/egosha7/go-loyalty-program.git/internal/helpers"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"time"
)

func BalanceHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, logger *zap.Logger) {
	// Извлечение имени пользователя из куки
	cookie, err := r.Cookie("auth")
	if err != nil {
		logger.Error("Пользователь не авторизован", zap.Error(err))
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}
	username := cookie.Value

	// Запрос баланса пользователя
	var currentBalance float64
	var totalWithdrawn *float64

	// Запрос текущего баланса из таблицы loyalty_balance
	currentBalance, err = queries.SelectUserBalance(r.Context(), conn, username)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных1", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	// Запрос суммы снятых средств (withdrawn) из таблицы loyalty_withdrawals
	totalWithdrawn, err = queries.SelectTotalWithdrawn(r.Context(), conn, username)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных2", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	// Формирование ответа
	balanceData := struct {
		Current   float64  `json:"current"`
		Withdrawn *float64 `json:"withdrawn"`
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
}

func WithdrawHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, logger *zap.Logger) {
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
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	err = json.NewDecoder(r.Body).Decode(&withdrawRequest)
	if err != nil {
		logger.Error("Неверный формат запроса", zap.Error(err))
		http.Error(w, "Неверный формат запроса", http.StatusUnprocessableEntity)
		return
	}

	// Проверка формата номера заказа
	if !regexp.MustCompile(`^\d+$`).MatchString(withdrawRequest.Order) || !helpers.IsLuhnValid(withdrawRequest.Order) {
		logger.Error("Неверный формат номера заказа", zap.String("order_number", withdrawRequest.Order))
		http.Error(w, "Неверный формат номера заказа", http.StatusUnprocessableEntity)
		return
	}

	// Проверка на наличие достаточного баланса для списания
	var currentBalance float64
	currentBalance, err = queries.SelectUserBalance(r.Context(), conn, username)
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

	// Проверка на существование заказа
	var orderExists bool
	orderExists, err = queries.CheckOrderExists(r.Context(), conn, withdrawRequest.Order)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	if orderExists {
		// Изменение заказа в таблице orders
		err = queries.UpdateOrder(r.Context(), conn, withdrawRequest.Order, withdrawRequest.Sum)
		if err != nil {
			logger.Error("Ошибка при обновлении заказа в базе данных", zap.Error(err))
			http.Error(w, "Ошибка при обновлении заказа в базе данных", http.StatusInternalServerError)
			return
		}
	} else {
		// Вставка нового заказа в таблицу orders
		err = queries.InsertNewOrder(r.Context(), conn, withdrawRequest.Order, username)
		if err != nil {
			logger.Error("Ошибка при добавлении заказа в базу данных", zap.Error(err))
			http.Error(w, "Ошибка при добавлении заказа в базу данных", http.StatusInternalServerError)
			return
		}
	}

	// Вставка данных о списании в таблицу loyalty_withdrawals
	err = queries.InsertWithdrawalData(r.Context(), conn, username, withdrawRequest.Order, withdrawRequest.Sum)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	// Обновление баланса пользователя
	err = queries.UpdateUserBalance(r.Context(), conn, username, withdrawRequest.Sum)
	if err != nil {
		logger.Error("Ошибка при обновлении баланса пользователя", zap.Error(err))
		http.Error(w, "Ошибка при обновлении баланса пользователя", http.StatusInternalServerError)
		return
	}

	// Отправка успешного ответа
	w.WriteHeader(http.StatusOK)
}

func WithdrawalsHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, logger *zap.Logger) {
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
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}

	// Итерация по результатам запроса и добавление их в список withdrawals
	for rows.Next() {
		var orderID string
		var withdrawnPoints float64
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
			Order       string  `json:"order"`
			Sum         float64 `json:"sum"`
			ProcessedAt string  `json:"processed_at"`
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
}
