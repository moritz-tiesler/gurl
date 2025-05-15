package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	"gurl/repository"
	"gurl/repository/tutorial"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

type URLRepo interface {
	*tutorial.Queries
}

func run() error {
	ctx := context.Background()

	queries, err := repository.New(ctx, "./gurl.db", ddl)
	if err != nil {
		return err
	}
	// list all authors
	urls, err := queries.ListUrls(ctx)
	if err != nil {
		return err
	}
	log.Println(urls)

	// create an author
	insertedUrl, err := queries.CreateUrl(ctx, tutorial.CreateUrlParams{
		Original: "https://www.zeit.de",
		Short:    "gurl.me/abba",
	})
	if err != nil {
		return err
	}
	log.Println(insertedUrl)

	// get the author we just inserted
	fetchedAuthor, err := queries.GetUrl(ctx, insertedUrl.ID)
	if err != nil {
		return err
	}

	// prints true
	log.Println(reflect.DeepEqual(insertedUrl, fetchedAuthor))
	return nil
}

func NewServer() *http.Server {
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(http.Dir("./static")))

	router.HandleFunc("POST /url", postURL)
	router.HandleFunc("GET /url", getURL)

	stack := Stack(
		LogRequestMiddleware(log.Printf),
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: stack(router),
	}

	return server
}

var shortCounter int
var counterMut sync.Mutex

var m map[string]string = make(map[string]string)
var mMut sync.RWMutex

func postURL(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	longURL := r.Form.Get("long_url")
	if longURL == "" {
		http.Error(w, "Missing form data", http.StatusBadRequest)
	}

	counterMut.Lock()
	shortURLKey := fmt.Sprintf("short%d", shortCounter)
	shortCounter++
	counterMut.Unlock()

	mMut.Lock()
	m[shortURLKey] = longURL
	mMut.Unlock()

	w.Write([]byte(fmt.Sprintf("localhost:8080/url/%s", shortURLKey)))
}

func getURL(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.zeit.de", http.StatusMovedPermanently)
}

func main() {
	// if err := run(); err != nil {
	// 	log.Fatal(err)

	// }

	s := NewServer()

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
