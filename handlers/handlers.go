package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"gurl/repository/tutorial"
	"gurl/wordgen"
	"net/http"
	"net/url"
	"sync"

	"github.com/pingcap/log"
)

type Entry struct {
	value tutorial.Url
}

type Cache struct {
	sync.RWMutex
	data map[string]Entry
}

func (c *Cache) Set(key string, value tutorial.Url) {
	c.Lock()
	defer c.Unlock()

	c.data[key] = Entry{value}
}

func (c *Cache) Get(key string) (tutorial.Url, bool) {
	c.RLock()
	defer c.RUnlock()
	e, ok := c.data[key]
	return e.value, ok
}

var c *Cache = &Cache{data: make(map[string]Entry)}

type Handler struct {
	DB        *tutorial.Queries
	Cache     *Cache
	Generator wordgen.NameGen
}

func New(q *tutorial.Queries) *Handler {
	return &Handler{
		DB:        q,
		Cache:     c,
		Generator: *wordgen.New(),
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
	}
	_, err = url.ParseRequestURI(longURL)
	if err != nil {
		http.Error(w, "Missing form data", http.StatusBadRequest)
	}

	entry, err := h.DB.CreateUrl(r.Context(), tutorial.CreateUrlParams{
		Original: longURL,
		Short:    "",
	})
	shortURLKey := h.Generator.Generate(int32(entry.ID))
	h.DB.UpdateUrl(r.Context(), tutorial.UpdateUrlParams{
		Short: shortURLKey,
		ID:    entry.ID,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	url := ""
	if r.TLS != nil {
		url += "https://"
	}
	url += "%s/url/%s"
	w.Write(fmt.Appendf(nil, url, r.Host, shortURLKey))
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

	url, err = h.DB.GetUrlByShortUrl(r.Context(), short)
	if err != nil {
		log.Error(err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Cache.Set(short, url)
	http.Redirect(w, r, url.Original, http.StatusFound)
}
