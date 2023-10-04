package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// User структура для представления пользователя
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, logger *zap.Logger) {
	// Парсинг JSON-данных из запроса
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		logger.Error("Ошибка при разборе JSON", zap.Error(err))
		http.Error(w, "Ошибка при разборе JSON", http.StatusBadRequest)
		return
	}

	// Отладочный лог: вывод данных о полученном пользователе
	logger.Debug("Получен запрос на регистрацию пользователя", zap.Any("user", user))

	// Проверка наличия обязательных полей в запросе
	if user.Login == "" || user.Password == "" {
		logger.Error("Неверный формат запроса: отсутствуют обязательные поля")
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Проверка, что логин уникальный
	var existingUser string
	err = conn.QueryRow(r.Context(), "SELECT login FROM users WHERE login = $1", user.Login).Scan(&existingUser)
	if err != nil && err != pgx.ErrNoRows {
		logger.Error("Ошибка при выполнении запроса к базе данных", zap.Error(err))
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}
	if existingUser != "" {
		logger.Info("Пользователь с таким логином уже существует", zap.String("login", user.Login))
		http.Error(w, "Пользователь с таким логином уже существует", http.StatusConflict)
		return
	}

	// Хэширование пароля перед сохранением
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Ошибка хэширования пароля", zap.Error(err))
		http.Error(w, "Ошибка хэширования пароля", http.StatusInternalServerError)
		return
	}

	// Вставка нового пользователя в таблицу users и получение user_id
	var userID int
	err = conn.QueryRow(r.Context(), "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id", user.Login, hashedPassword).Scan(&userID)
	if err != nil {
		// Обработка ошибки
		http.Error(w, "Ошибка при регистрации пользователя", http.StatusInternalServerError)
		return
	}

	// Вставка записи в таблицу loyalty_balance с начальными баллами (0.0)
	_, err = conn.Exec(r.Context(), "INSERT INTO loyalty_balance (user_id, points) VALUES ($1, $2)", userID, 0.0)
	if err != nil {
		// Обработка ошибки
		http.Error(w, "Ошибка при добавлении начальных баллов в loyalty_balance", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    user.Login, // Используйте логин пользователя как идентификатор
		HttpOnly: true,
	})

	// Ответ клиенту
	w.WriteHeader(http.StatusOK)
	logger.Info("Пользователь успешно зарегистрирован", zap.String("login", user.Login))
	fmt.Fprintf(w, "Пользователь %s успешно аутентифицирован", user.Login)
}

func LoginUser(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	// Парсинг JSON-данных из запроса
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Ошибка при разборе JSON", http.StatusBadRequest)
		return
	}

	// Проверка наличия обязательных полей в запросе
	if user.Login == "" || user.Password == "" {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Получение хэшированного пароля из базы данных по логину
	var hashedPassword string
	err = conn.QueryRow(r.Context(), "SELECT password FROM users WHERE login = $1", user.Login).Scan(&hashedPassword)
	if err == pgx.ErrNoRows {
		http.Error(w, "Неверная пара логин/пароль", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Неверная пара логин/пароль", http.StatusUnauthorized)
		return
	}

	// Установка куки
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    user.Login, // Используйте логин пользователя как идентификатор
		HttpOnly: true,
	})

	// Ответ клиенту
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Пользователь %s успешно аутентифицирован", user.Login)
}
