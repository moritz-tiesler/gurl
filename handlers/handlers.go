package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gurl/lru_cache"
	"gurl/repository/tutorial"
	"gurl/templates"
	"gurl/wordgen"
	"net/http"
	"net/url"
	"strings"

	"github.com/pingcap/log"
)

type Handler struct {
	Repo      Repo
	Cache     *lru_cache.Cache[string, tutorial.Url]
	Generator *wordgen.NameGen
}

type Repo interface {
	CreateUrl(context.Context, tutorial.CreateUrlParams) (tutorial.Url, error)
	GetUrlByShortUrl(ctx context.Context, short string) (tutorial.Url, error)
	UpdateUrl(ctx context.Context, arg tutorial.UpdateUrlParams) error
}

func New(repo Repo) *Handler {
	return &Handler{
		Repo:      repo,
		Cache:     lru_cache.New[string, tutorial.Url](1024),
		Generator: wordgen.New(),
	}
}

func (h *Handler) PostURL(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	longURL := r.Form.Get("long_url")
	if longURL == "" {
		http.Error(w, "Missing form data", http.StatusBadRequest)
		return
	}

	_, err = url.ParseRequestURI(longURL)
	if err != nil {
		http.Error(w, "Missing form data", http.StatusBadRequest)
		return
	}

	entry, err := h.Repo.CreateUrl(r.Context(), tutorial.CreateUrlParams{
		Original: longURL,
		Short:    "",
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortURLKey := h.Generator.Generate(int32(entry.ID))
	err = h.Repo.UpdateUrl(r.Context(), tutorial.UpdateUrlParams{
		Short: shortURLKey,
		ID:    entry.ID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Cache.Add(shortURLKey, entry)

	url := ""
	if !strings.HasPrefix(r.Host, "localhost") {
		url += "https://"
	}
	url += fmt.Sprintf("%s/url/%s", r.Host, shortURLKey)

	t := templates.URL{Value: url}
	html := t.Render()

	// TODO: answer with html input element, makes for nicer styling
	// use templating
	w.Write(html)
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	var url tutorial.Url
	var err error

	short := r.PathValue("short")
	url, ok := h.Cache.Get(short)
	if ok {
		http.Redirect(w, r, url.Original, http.StatusFound)
		return
	}

	url, err = h.Repo.GetUrlByShortUrl(r.Context(), short)
	if err != nil {
		log.Error(err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Cache.Add(short, url)
	http.Redirect(w, r, url.Original, http.StatusFound)
}
