package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/egosha7/go-loyalty-program.git/internal/config"
	"github.com/egosha7/go-loyalty-program.git/internal/helpers"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func sendRequestToAccrualSystem(orderNumber string, cfg *config.Config) (*AccrualResponse, error) {
	// Формируйте URL для запроса к системе расчёта баллов на основе cfg.AccrualSystemAddr и orderNumber
	url := fmt.Sprintf("%s/api/orders/%s", cfg.AccrualSystemAddr, orderNumber)

	// Отправьте GET-запрос к системе расчёта баллов
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Обработайте ответ от системы расчёта баллов
	if resp.StatusCode == http.StatusOK {
		var accrualResponse AccrualResponse
		err := json.NewDecoder(resp.Body).Decode(&accrualResponse)
		if err != nil {
			return nil, err
		}
		return &accrualResponse, nil
	} else if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	} else {
		return nil, fmt.Errorf("ошибка при запросе к системе расчёта баллов: %d", resp.StatusCode)
	}
}

type Order struct {
	Number     string    `json:"number"`
	Status     *string   `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func OrdersHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, cfg *config.Config, logger *zap.Logger) {
	// Проверка аутентификации пользователя
	cookie, err := r.Cookie("auth")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}
	user := cookie.Value

	var userID int
	err = conn.QueryRow(r.Context(), "SELECT user_id FROM users WHERE login = $1", user).Scan(&userID)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		logger.Error("Неверный формат запроса", zap.String("content_type", contentType))
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Ошибка при чтении тела запроса", zap.Error(err))
		http.Error(w, "Ошибка при чтении тела запроса", http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(bodyBytes))

	// Проверка формата номера заказа
	if !regexp.MustCompile(`^\d+$`).MatchString(orderNumber) || !helpers.IsLuhnValid(orderNumber) {
		logger.Error("Неверный формат номера заказа", zap.String("order_number", orderNumber))
		http.Error(w, "Неверный формат номера заказа", http.StatusUnprocessableEntity)
		return
	}

	// Проверка уникальности номера заказа для данного пользователя
	var existingUser int
	err = conn.QueryRow(r.Context(), "SELECT user_id FROM orders WHERE order_number = $1", orderNumber).Scan(&existingUser)
	if err != nil && err != pgx.ErrNoRows {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}
	if existingUser != userID && existingUser != 0 {
		logger.Error("Номер заказа уже был загружен другим пользователем")
		http.Error(w, "Номер заказа уже был загружен другим пользователем", http.StatusConflict)
		return
	} else if existingUser == userID {
		logger.Info("Номер заказа уже был загружен этим пользователем", zap.String("order_number", orderNumber))
		http.Error(w, "Номер заказа уже был загружен этим пользователем", http.StatusOK)
		return
	}

	accrualResponse, err := sendRequestToAccrualSystem(orderNumber, cfg)
	if err != nil {
		logger.Error("Ошибка при запросе к системе расчёта баллов", zap.Error(err))
		http.Error(w, "Ошибка при запросе к системе расчёта баллов", http.StatusInternalServerError)
		return
	}

	currentTime := time.Now()

	if accrualResponse != nil {
		// Используем accrualResponse если он не равен nil
		_, err = conn.Exec(r.Context(), "INSERT INTO orders (order_number, user_id, order_status, timestamp, order_accural) VALUES ($1, $2, $3, $4, $5)",
			orderNumber, userID, accrualResponse.Status, currentTime.Format(time.RFC3339), accrualResponse.Accrual)
		if err != nil {
			logger.Error("Ошибка при добавлении номера заказа в базу данных", zap.Error(err))
			http.Error(w, "Ошибка при добавлении номера заказа в базу данных", http.StatusInternalServerError)
			return
		}

		_, err = conn.Exec(r.Context(), "UPDATE loyalty_balance SET points = points + $1 WHERE user_id = $2",
			accrualResponse.Accrual, userID)
		if err != nil {
			logger.Error("Ошибка при обновлении баланса", zap.Error(err))
			http.Error(w, "Ошибка при обновлении баланса", http.StatusInternalServerError)
			return
		}
	} else {
		// Не используем accrualResponse если он равен nil
		_, err = conn.Exec(r.Context(), "INSERT INTO orders (order_number, user_id, order_status, timestamp) VALUES ($1, $2, $3, $4)",
			orderNumber, userID, "NEW", currentTime.Format(time.RFC3339))
		if err != nil {
			logger.Error("Ошибка при добавлении номера заказа в базу данных", zap.Error(err))
			http.Error(w, "Ошибка при добавлении номера заказа в базу данных", http.StatusInternalServerError)
			return
		}
	}

	// Ответ клиенту
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Номер заказа %s принят в обработку", orderNumber)
}

func OrdersListHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, logger *zap.Logger) {
	// Извлечение имени пользователя из куки
	cookie, err := r.Cookie("auth")
	if err != nil {
		logger.Error("Пользователь не авторизован", zap.Error(err))
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}
	username := cookie.Value

	// Запрос списка заказов пользователя
	rows, err := conn.Query(r.Context(), "SELECT order_number, order_status, order_accural, timestamp FROM orders  WHERE user_id = (SELECT user_id FROM users WHERE login = $1) ORDER BY timestamp ASC", username)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	orders := []Order{}

	for rows.Next() {
		var order Order
		var accrual sql.NullFloat64 // Используем sql.NullFloat64 для обработки NULL-значений

		if err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt); err != nil {
			logger.Error("Ошибка при сканировании результатов запроса", zap.Error(err))
			http.Error(w, "Ошибка при сканировании результатов запроса", http.StatusInternalServerError)
			return
		}

		// Проверяем, есть ли значение accrual
		if accrual.Valid {
			order.Accrual = accrual.Float64
		} else {
			order.Accrual = 0 // Или другое значение по умолчанию
		}

		orders = append(orders, order)
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Отправка ответа клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		logger.Error("Ошибка при кодировании ответа", zap.Error(err))
		http.Error(w, "Ошибка при кодировании ответа", http.StatusInternalServerError)
		return
	}
}
