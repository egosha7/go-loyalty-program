package compress

import (
	"compress/gzip"
	"net/http"
)

type GzipMiddleware struct{}

func (m *GzipMiddleware) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(`Content-Encoding`) == `gzip` {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer gz.Close()

				// Замена тела запроса на распакованное содержимое
				r.Body = http.MaxBytesReader(w, gz, r.ContentLength)
				r.Header.Del("Content-Encoding")
				r.Header.Del("Content-Length")
			}

			// Передаем управление следующему обработчику
			next.ServeHTTP(w, r)
		},
	)
}
