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
	// Запрос для получения списка таблиц
	query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`

	// Выполнение запроса и получение результатов
	rows, err := conn.Query(r.Context(), query)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса", zap.Error(err))
		return
	}
	defer rows.Close()
	// Считывание и вывод списка таблиц
	var tableName string
	for rows.Next() {
		if err := rows.Scan(&tableName); err != nil {
			logger.Error("Ошибка при сканировании строки результата", zap.Error(err))
			return
		}
		logger.Info("Найдена таблица", zap.String("table_name", tableName))
	}

	if err := rows.Err(); err != nil {
		logger.Error("Ошибка при чтении строк результата", zap.Error(err))
		return
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
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

	// Вставка нового пользователя в базу данных
	_, err = conn.Exec(r.Context(), "INSERT INTO users (login, password) VALUES ($1, $2)", user.Login, hashedPassword)
	if err != nil {
		logger.Error("Ошибка при добавлении пользователя в базу данных", zap.Error(err))
		http.Error(w, "Ошибка при добавлении пользователя в базу данных", http.StatusInternalServerError)
		return
	}

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

	// В случае успешной аутентификации вы можете отправить куки или токен для аутентифицированного пользователя
	// Ниже приведен пример установки куки
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    user.Login, // Используйте логин пользователя как идентификатор
		HttpOnly: true,
	})

	// Ответ клиенту
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Пользователь %s успешно аутентифицирован", user.Login)
}
