package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"
)

type UrlStore interface {
	BeginTx() (*sql.Tx, error)

	InsertURLToDb(tx *sql.Tx, url string) (int, error)
	UpdateHashToDb(tx *sql.Tx, id int, hash string) error
	GetURLFromDb(hash string) (string, time.Time, error)
}

type CacheStore interface {
	Get(hash string) (string, bool)
	Put(hash, url string, expire time.Time)
}

type Handler struct {
	store  UrlStore
	config Config
	cache  CacheStore
}

func New(s UrlStore, config Config, cache CacheStore) *Handler {
	return &Handler{
		store:  s,
		config: config,
		cache:  cache,
	}
}

func isValidURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var ur CreateShortURLRequest

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&ur)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !isValidURL(ur.URL) {
		http.Error(w, "invalid url: must include http:// or https://", http.StatusBadRequest)
		return
	}

	start := time.Now()

	tx, err := h.store.BeginTx()

	log.Println("begin took:", time.Since(start))

	if err != nil {
		log.Println("start transaction fail :", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	t2 := time.Now()
	id, err := h.store.InsertURLToDb(tx, ur.URL)
	log.Println("insert took:", time.Since(t2))

	if err != nil {
		log.Println("insert error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return

	}

	hash := EncodeBase62(id, h.config.MixMultiplierSecret)

	t3 := time.Now()

	if err := h.store.UpdateHashToDb(tx, id, hash); err != nil {
		log.Println("update error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	log.Println("update took:", time.Since(t3))

	t4 := time.Now()

	err = tx.Commit()
	log.Println("commit took:", time.Since(t4))

	if err != nil {
		log.Println("update error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return

	}

	shortURL := h.config.SITE_URL + "/" + hash

	json.NewEncoder(w).Encode(map[string]string{"short_url": shortURL})
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")

	// 1. Check cache first
	if url, ok := h.cache.Get(hash); ok {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		return
	}

	// 2. Cache miss -> query DB
	url, expire, err := h.store.GetURLFromDb(hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		log.Println("get url error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	// 3. Check expiration
	if time.Now().After(expire) {
		http.Error(w, "URL has expired", http.StatusGone)
		return
	}
	// 4. Store in cache
	h.cache.Put(hash, url, expire)

	// 5. Redirect
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
