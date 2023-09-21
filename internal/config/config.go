package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net"
)

// Config - структура конфигурации приложения
type Config struct {
	Addr              string `env:"RUN_ADDRESS"`            // Адрес сервера
	DataBaseURI       string `env:"DATABASE_URI"`           // Адрес базы данных
	AccrualSystemAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"` // Адрес системы расчёта начислений
}

// Default - функция для создания новой конфигурации с значениями по умолчанию
func Default() *Config {
	return &Config{
		Addr:              "localhost:8080",
		DataBaseURI:       "postgres://postgres:egosha@localhost:5432/loyaltydb",
		AccrualSystemAddr: "http://localhost:8081",
	}
}

// OnFlag - функция для чтения значений из флагов командной строки и записи их в структуру Config
func OnFlag(logger *zap.Logger) *Config {
	defaultValue := Default()

	// Инициализация флагов командной строки
	config := Config{}
	flag.StringVar(&config.Addr, "a", defaultValue.Addr, "HTTP-адрес сервера")
	flag.StringVar(&config.DataBaseURI, "d", defaultValue.DataBaseURI, "Адрес базы данных")
	flag.StringVar(&config.AccrualSystemAddr, "r", defaultValue.AccrualSystemAddr, "Адрес системы расчёта начислений")
	flag.Parse()

	// Загрузка переменных окружения из файла .env, если он есть
	godotenv.Load()

	// Парсинг переменных окружения в структуру Config
	if err := env.Parse(&config); err != nil {
		logger.Error("Ошибка при парсинге переменных окружения", zap.Error(err))
	}

	// Проверка корректности введенных значений флагов
	if _, _, err := net.SplitHostPort(config.Addr); err != nil {
		panic(err)
	}
	if _, _, err := net.SplitHostPort(config.AccrualSystemAddr); err != nil {
		panic(err)
	}

	return &config
}
