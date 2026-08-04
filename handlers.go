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

type URLStore interface {
	BeginTx() (*sql.Tx, error)

	InsertURLToDB(tx *sql.Tx, url string) (int, error)
	UpdateHashToDB(tx *sql.Tx, id int, hash string) error
	GetURLFromDB(hash string) (string, time.Time, error)

	UpdateLastAccessTimes(nodes []DirtyNode) error
}

type CacheStore interface {
	Get(hash string) (string, bool)
	Put(hash, url string, expire time.Time, dirty bool)
	GetDirtyNodes() []DirtyNode
	MarkClean(nodes []DirtyNode)
	Touch(hash string)
}

type LastSync interface {
	Set(time.Time)
	Get() time.Time
	CheckInterval(time.Time) int
	TryStartSync() bool
	FinishSync()
}

type Handler struct {
	store         URLStore
	config        Config
	cache         CacheStore
	lastSyncState LastSync
}

func New(s URLStore, config Config, cache CacheStore, lastSyncState LastSync) *Handler {
	return &Handler{
		store:         s,
		config:        config,
		cache:         cache,
		lastSyncState: lastSyncState,
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
	if err := decoder.Decode(&ur); err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !isValidURL(ur.URL) {
		http.Error(w, "invalid url: must include http:// or https://", http.StatusBadRequest)
		return
	}

	tx, err := h.store.BeginTx()
	if err != nil {
		log.Println("start transaction fail:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	id, err := h.store.InsertURLToDB(tx, ur.URL)
	if err != nil {
		log.Println("insert error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	hash := EncodeBase62(id, h.config.MixMultiplierSecret)

	if err := h.store.UpdateHashToDB(tx, id, hash); err != nil {
		log.Println("update error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println("commit error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	shortURL := h.config.SITE_URL + "/" + hash

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"short_url": shortURL}); err != nil {
		log.Println("response encode error:", err)
	}
}

func (h *Handler) SyncLastAccessTime() {
	dirtyNodes := h.cache.GetDirtyNodes()

	if len(dirtyNodes) == 0 {
		return
	}

	if err := h.store.UpdateLastAccessTimes(dirtyNodes); err != nil {
		log.Println("sync failed:", err)
		return
	}

	h.cache.MarkClean(dirtyNodes)
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if h.lastSyncState.TryStartSync() {
		go func() {
			defer h.lastSyncState.FinishSync()
			h.SyncLastAccessTime()
		}()
	}

	hash := r.PathValue("hash")

	// Cache hit — Get() already updates lastAccessTime/dirty/MRU internally
	if url, ok := h.cache.Get(hash); ok {
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	// Cache miss
	url, expire, err := h.store.GetURLFromDB(hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		log.Println("get url error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if time.Now().After(expire) {
		http.Error(w, "URL has expired", http.StatusGone)
		return
	}

	// Load into cache, already synchronized with DB
	h.cache.Put(hash, url, expire, false)

	// This request accessed it
	h.cache.Touch(hash)

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
