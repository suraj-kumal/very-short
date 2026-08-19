package middleware

import "net/http"

type Limiter struct {
	sem chan struct{}
}

func NewLimiter(limit int) *Limiter {
	return &Limiter{
		sem: make(chan struct{}, limit),
	}
}

func (l *Limiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case l.sem <- struct{}{}:
			defer func() {
				<-l.sem
			}()

			next.ServeHTTP(w, r)

		default:
			http.Error(
				w,
				"too many requests",
				http.StatusTooManyRequests,
			)
		}
	})
}
