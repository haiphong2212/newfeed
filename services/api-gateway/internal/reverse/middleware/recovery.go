package middleware

import (
	"net/http"

	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/logger"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/response"
)

func Recovery(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					log.Error("panic recovered", map[string]any{
						"request_id": RequestIDFromContext(r.Context()),
						"panic":      recovered,
					})
					response.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
