package routes

import (
	"github.com/egosha7/go-loyalty-program.git/api/handlers"
	"github.com/egosha7/go-loyalty-program.git/internal/compress"
	"github.com/egosha7/go-loyalty-program.git/internal/config"
	"github.com/egosha7/go-loyalty-program.git/internal/db"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"net/http"
)

func SetupRoutes(cfg *config.Config, conn *pgx.Conn, logger *zap.Logger) http.Handler {

	// Создание роутера
	r := chi.NewRouter()

	gzipMiddleware := compress.GzipMiddleware{}

	// Создание группы роутера
	r.Group(
		func(route chi.Router) {
			route.Use(gzipMiddleware.Apply)

			route.Get(
				"/ping", func(w http.ResponseWriter, r *http.Request) {
					db.PingDB(w, r, conn)
				},
			)

			// Роут для регистрации пользователя
			route.Post(
				"/api/user/register", func(w http.ResponseWriter, r *http.Request) {
					handlers.RegisterUser(w, r, conn)
				},
			)

			// Роут для аутентификации пользователя
			route.Post(
				"/api/user/login", func(w http.ResponseWriter, r *http.Request) {
					handlers.LoginUser(w, r, conn)
				},
			)

			// Роут для отправки номера заказа
			route.Post(
				"/api/user/orders", func(w http.ResponseWriter, r *http.Request) {
					handlers.OrdersHandler(w, r, conn, cfg)
				},
			)

			// Роут для полученя информации о заказах
			route.Get(
				"/api/user/orders", func(w http.ResponseWriter, r *http.Request) {
					handlers.OrdersListHandler(w, r, conn)
				},
			)

			// Роут для полученя информации о текущем балансе
			route.Get(
				"/api/user/balance", func(w http.ResponseWriter, r *http.Request) {
					handlers.BalanceHandler(w, r, conn)
				},
			)

			// Роут для полученя информации о текущем балансе
			route.Post(
				"/api/user/balance/withdraw", func(w http.ResponseWriter, r *http.Request) {
					handlers.WithdrawHandler(w, r, conn)
				},
			)

			// Роут для полученя информации о текущем балансе
			route.Get(
				"/api/user/withdrawals", func(w http.ResponseWriter, r *http.Request) {
					handlers.WithdrawalsHandler(w, r, conn)
				},
			)

		},
	)

	return r
}
