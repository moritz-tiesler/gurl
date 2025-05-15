package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"time"

	"gurl/handlers"
	"gurl/repository"
	"gurl/repository/tutorial"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

type Server struct {
	*http.Server
	*tutorial.Queries
}

func NewServer(h *handlers.Handler) *http.Server {
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(http.Dir("./static")))

	router.HandleFunc("POST /url", h.PostURL)
	router.HandleFunc("GET /url/{short}", h.GetURL)

	stack := Stack(
		LogRequestMiddleware(log.Printf),
	)

	s := &http.Server{
		Addr:    ":8080",
		Handler: stack(router),
	}

	return s
}

func main() {
	ctx := context.Background()

	queries, err := repository.New(ctx, "./gurl.db", ddl)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}

	handler := handlers.New(queries)

	s := NewServer(handler)

	log.Printf("launching server at %v", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Could not launch server")
	}

}

type middleware func(http.Handler) http.Handler

func Stack(middlewares ...middleware) middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i > -1; i-- {
			m := middlewares[i]
			next = m(next)
		}
		return next
	}
}

func LogRequestMiddleware(loggingFunc func(string, ...any)) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loggingFunc("%v: LOG %s - %s %s %s\n", time.Now(), r.RemoteAddr, r.Proto, r.Method, r.URL)

			next.ServeHTTP(w, r)
		})
	}
}
