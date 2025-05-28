package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gurl/lru_cache"
	urlRepo "gurl/repository/url"
	"gurl/templates"
	"gurl/wordgen"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Handler struct {
	Repo      *urlRepo.Queries
	DB        *sql.DB
	Cache     *lru_cache.Cache[string, urlRepo.Url]
	Generator *wordgen.NameGen
}

func New(repo *urlRepo.Queries, db *sql.DB) *Handler {
	return &Handler{
		Repo:      repo,
		DB:        db,
		Cache:     lru_cache.New[string, urlRepo.Url](1024 * 8),
		Generator: wordgen.New(),
	}
}

var (
	ErrNotFound    error = errors.New("entity not found")
	ErrRequestData error = errors.New("cannot handle request data")
	ErrDatabase    error = errors.New("database error")
)

func (h *Handler) PostURL() func(w http.ResponseWriter, r *http.Request) {
	return MakeHandler(h.postURL)
}

func (h *Handler) GetURL() func(w http.ResponseWriter, r *http.Request) {
	return MakeHandler(h.getURL)
}

func (h *Handler) postURL(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("cannot parse form: %w", ErrRequestData)
	}

	longURL := r.Form.Get("long_url")
	if longURL == "" {
		return fmt.Errorf("missing form data: %w", ErrRequestData)
	}

	// TODO move url validity check to JS
	_, err = url.ParseRequestURI(longURL)
	if err != nil {
		return fmt.Errorf("cannot parse url: %w", ErrRequestData)
	}

	shortURLKey, err := h.saveURL(r.Context(), longURL)
	if err != nil {
		return fmt.Errorf("error calling saveURL(): %s: %w", err, ErrDatabase)
	}

	url := ""
	if !strings.HasPrefix(r.Host, "localhost") {
		url += "https://"
	}
	url += fmt.Sprintf("%s/url/%s", r.Host, shortURLKey)

	t := templates.URL{Value: url}
	html := t.Render()

	w.Write(html)
	return nil
}

func (h *Handler) getURL(w http.ResponseWriter, r *http.Request) error {
	var url urlRepo.Url
	var err error

	short := r.PathValue("short")
	if short == "" {
		return fmt.Errorf("empty url: %w", ErrRequestData)
	}
	url, ok := h.Cache.Get(short)
	if ok {
		http.Redirect(w, r, url.Original, http.StatusFound)
		return nil
	}

	url, err = h.Repo.GetUrlByShortUrl(r.Context(), short)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("'%s' %s: %w", short, err, ErrNotFound)
		}
		return fmt.Errorf("%s: %w", err, ErrDatabase)
	}

	h.Cache.Add(short, url)
	http.Redirect(w, r, url.Original, http.StatusFound)

	return nil
}

func MakeHandler(h func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			log.Printf("%v: LOG %s - %s %s %s %s\n", time.Now(), r.RemoteAddr, r.Proto, r.Method, r.URL, err)
			if errors.Is(err, ErrNotFound) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			if errors.Is(err, ErrRequestData) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) saveURL(ctx context.Context, longURL string) (string, error) {
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("BeginTx() %s", err)
	}
	defer tx.Rollback()
	qtx := h.Repo.WithTx(tx)

	entry, err := qtx.CreateUrl(ctx, urlRepo.CreateUrlParams{
		Original: longURL,
		Short:    "",
	})

	if err != nil {
		return "", fmt.Errorf("CreateUrl() %s", err)
	}

	shortURLKey := h.Generator.Generate(int32(entry.ID))
	err = qtx.UpdateUrl(ctx, urlRepo.UpdateUrlParams{
		Short: shortURLKey,
		ID:    entry.ID,
	})
	if err != nil {
		return "", fmt.Errorf("UpdateUrl() %s", err)
	}
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("Commit() %s", err)
	}
	return shortURLKey, nil
}
