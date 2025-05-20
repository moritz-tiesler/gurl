package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"gurl/lru_cache"
	"gurl/repository"
	urlRepo "gurl/repository/url"
	"gurl/templates"
	"gurl/wordgen"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	Repo      repository.Repo
	Cache     *lru_cache.Cache[string, urlRepo.Url]
	Generator *wordgen.NameGen
}

func New(repo repository.Repo) *Handler {
	return &Handler{
		Repo:      repo,
		Cache:     lru_cache.New[string, urlRepo.Url](1024 * 8),
		Generator: wordgen.New(),
	}
}

func (h *Handler) PostURL(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("%s\n", err.Error())
		http.Error(w, "invalid form", http.StatusInternalServerError)
		return
	}

	longURL := r.Form.Get("long_url")
	if longURL == "" {
		http.Error(w, "Missing form data", http.StatusBadRequest)
		return
	}

	// TODO move url validity check to JS
	_, err = url.ParseRequestURI(longURL)
	if err != nil {
		log.Printf("%s\n", err.Error())
		http.Error(w, "Missing form data", http.StatusBadRequest)
		return
	}

	tx, err := h.Repo.DB().BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("%s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	qtx := h.Repo.WithTx(tx)

	entry, err := qtx.CreateUrl(r.Context(), urlRepo.CreateUrlParams{
		Original: longURL,
		Short:    "",
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\n", err.Error())
		return
	}

	shortURLKey := h.Generator.Generate(int32(entry.ID))
	err = qtx.UpdateUrl(r.Context(), urlRepo.UpdateUrlParams{
		Short: shortURLKey,
		ID:    entry.ID,
	})
	if err != nil {
		log.Printf("%s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\n", err.Error())
	}

	url := ""
	if !strings.HasPrefix(r.Host, "localhost") {
		url += "https://"
	}
	url += fmt.Sprintf("%s/url/%s", r.Host, shortURLKey)

	t := templates.URL{Value: url}
	html := t.Render()

	w.Write(html)
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	var url urlRepo.Url
	var err error

	short := r.PathValue("short")
	log.Printf("requested=%s\n", short)
	url, ok := h.Cache.Get(short)
	if ok {
		log.Printf("cache hit=%s\n", url.Original)
		http.Redirect(w, r, url.Original, http.StatusFound)
		return
	}

	url, err = h.Repo.GetUrlByShortUrl(r.Context(), short)
	if err != nil {
		log.Printf("%s\n", err)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("db hit=%s\n", url.Original)
	h.Cache.Add(short, url)
	http.Redirect(w, r, url.Original, http.StatusFound)
}
