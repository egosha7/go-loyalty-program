package routes

import (
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

		},
	)

	return r
}
