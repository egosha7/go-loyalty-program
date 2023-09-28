package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// User структура для представления пользователя
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
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

	// Проверка, что логин уникальный
	var existingUser string
	err = conn.QueryRow(r.Context(), "SELECT login FROM users WHERE login = $1", user.Login).Scan(&existingUser)
	if err != nil && err != pgx.ErrNoRows {
		http.Error(w, "Ошибка при выполнении запроса к базе данных", http.StatusInternalServerError)
		return
	}
	if existingUser != "" {
		http.Error(w, "Пользователь с таким логином уже существует", http.StatusConflict)
		return
	}

	// Хэширование пароля перед сохранением
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка хэширования пароля", http.StatusInternalServerError)
		return
	}

	// Вставка нового пользователя в базу данных
	_, err = conn.Exec(r.Context(), "INSERT INTO users (login, password) VALUES ($1, $2)", user.Login, hashedPassword)
	if err != nil {
		http.Error(w, "Ошибка при добавлении пользователя в базу данных", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Пользователь %s успешно зарегистрирован", user.Login)
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
