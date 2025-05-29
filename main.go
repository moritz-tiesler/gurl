package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"time"

	"gurl/handlers"

	"gurl/repository/url"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

type Server struct {
	*http.Server
}

func NewServer(mw middleware, mux *http.ServeMux) *http.Server {
	s := &http.Server{
		Addr:    ":8080",
		Handler: mw(mux),
	}
	return s
}

func main() {
	ctx := context.Background()

	db, err := sql.Open("sqlite", "./gurl.db")
	if err != nil {
		log.Fatal("Could not opn db connection")
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatalf("Failed to set journal_mode: %v", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		log.Fatalf("Could not create scheme: %s", err)
	}

	queries := url.New(db)
	handler := handlers.New(queries, db)

	router := http.NewServeMux()
	setupRoutes(handler, router)

	middlewareStack := Stack(
		LogRequestMiddleware(log.Printf),
	)

	s := NewServer(middlewareStack, router)

	log.Printf("launching server at %v", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Could not launch server")
	}

}
func setupRoutes(h *handlers.Handler, router *http.ServeMux) {
	router.Handle("/", http.FileServer(http.Dir("./static")))

	router.HandleFunc("POST /url", h.PostURL())
	router.HandleFunc("GET /url/{short}", h.GetURL())
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

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (ww *wrappedWriter) WriteHeader(statusCode int) {
	ww.ResponseWriter.WriteHeader(statusCode)
	ww.statusCode = statusCode
}

func LogRequestMiddleware(loggingFunc func(string, ...any)) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loggingFunc("%v: LOG %s - %s %s %s\n", time.Now(), r.RemoteAddr, r.Proto, r.Method, r.URL)

			wrapped := &wrappedWriter{w, http.StatusOK}
			next.ServeHTTP(wrapped, r)
			loggingFunc("STATUS: %v", wrapped.statusCode)
		})
	}
}
