package handlers

import (
	"fmt"
	"gurl/repository/tutorial"
	"net/http"
	"net/url"
	"sync"
)

type Handler struct {
	DB *tutorial.Queries
}

var shortCounter int
var counterMut sync.Mutex

var m map[string]string = make(map[string]string)
var mMut sync.RWMutex

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

	counterMut.Lock()
	shortURLKey := fmt.Sprintf("short%d", shortCounter)
	shortCounter++
	counterMut.Unlock()
	_, err = h.DB.CreateUrl(r.Context(), tutorial.CreateUrlParams{
		Original: longURL,
		Short:    shortURLKey,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(fmt.Appendf(nil, "localhost:8080/url/%s", shortURLKey))
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	short := r.PathValue("short")
	long, err := h.DB.GetUrlByShortUrl(r.Context(), short)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, long.Original, http.StatusMovedPermanently)
}
