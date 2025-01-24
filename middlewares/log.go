package middlewares

import (
	"log"
	"net/http"
)

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Response)
		log.Printf("Method: %s, Path: %s, IP: %s", 
			r.Method, 
			r.URL.Path, 
			r.RemoteAddr,
		)
		next.ServeHTTP(w, r)
	})
}