package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/suraj-kumal/very-short/internal/contracts"
	"github.com/suraj-kumal/very-short/internal/encode"
	"github.com/suraj-kumal/very-short/internal/models"
)

type URLStore interface {
	BeginTx() (*sql.Tx, error)

	InsertURLToDB(tx *sql.Tx, url string) (int, error)
	UpdateHashToDB(tx *sql.Tx, id int, hash string) error
	GetURLFromDB(hash string) (string, time.Time, error)

	UpdateLastAccessTimes(nodes []contracts.DirtyNode) error
}

type CacheStore interface {
	Get(hash string) (string, bool)
	Put(hash, url string, expire time.Time, dirty bool)
	GetDirtyNodes() []contracts.DirtyNode
	MarkClean(nodes []contracts.DirtyNode)
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
	store               URLStore
	SiteURL             string
	MixMultiplierSecret int
	cache               CacheStore
	lastSyncState       LastSync
}

func HandlerStore(s URLStore, SiteURL string, MixMultiplierSecret int, cache CacheStore, lastSyncState LastSync) *Handler {
	return &Handler{
		store:               s,
		SiteURL:             SiteURL,
		MixMultiplierSecret: MixMultiplierSecret,
		cache:               cache,
		lastSyncState:       lastSyncState,
	}
}

func isValidURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

var (
	resultSuccessTmpl = template.Must(template.New("success").Parse(`
		<div class="mono text-sm">
			<div class="flex items-center justify-between border-b hairline pb-3 mb-3">
				<span class="index-num">STATUS : </span>
				<span style="color:var(--ink)">CREATED</span>
			</div>
			<div class="flex items-center justify-between gap-3">
				<span class="text-lg break-all" style="color:var(--red)">{{.ShortURL}}</span>
				<button type="button"
					class="btn"
					style="padding:0.5rem 1rem; white-space:nowrap"
					onclick="copyShortURL(this, '{{.ShortURL}}')">
					Copy
				</button>
			</div>
		</div>`))

	resultErrorTmpl = template.Must(template.New("error").Parse(`
<div class="mono text-sm pl-3" style="border-left:2px solid var(--red); color:var(--ink-soft)">
	<span class="index-num" style="color:var(--red)">ERROR</span><br>
	{{.Message}}
</div>`))
)

func writeResult(w http.ResponseWriter, status int, tmpl *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = tmpl.Execute(w, data)
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeResult(w, http.StatusBadRequest, resultErrorTmpl, struct{ Message string }{"Invalid request body."})
		return
	}
	ur := models.CreateShortURLRequest{URL: r.FormValue("url")}
	if !isValidURL(ur.URL) {
		writeResult(w, http.StatusBadRequest, resultErrorTmpl, struct{ Message string }{"URL must include http:// or https://"})
		return
	}
	tx, err := h.store.BeginTx()
	if err != nil {
		log.Println("start transaction fail:", err)
		writeResult(w, http.StatusInternalServerError, resultErrorTmpl, struct{ Message string }{"Something went wrong."})
		return
	}
	defer tx.Rollback()

	id, err := h.store.InsertURLToDB(tx, ur.URL)
	if err != nil {
		log.Println("insert error:", err)
		writeResult(w, http.StatusInternalServerError, resultErrorTmpl, struct{ Message string }{"Something went wrong."})
		return
	}
	hash := encode.EncodeBase62(id, h.MixMultiplierSecret)
	if err := h.store.UpdateHashToDB(tx, id, hash); err != nil {
		log.Println("update error:", err)
		writeResult(w, http.StatusInternalServerError, resultErrorTmpl, struct{ Message string }{"Something went wrong."})
		return
	}
	if err := tx.Commit(); err != nil {
		log.Println("commit error:", err)
		writeResult(w, http.StatusInternalServerError, resultErrorTmpl, struct{ Message string }{"Something went wrong."})
		return
	}

	shortURL := h.SiteURL + "/" + hash
	writeResult(w, http.StatusCreated, resultSuccessTmpl, struct{ ShortURL string }{shortURL})
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
			http.ServeFile(w, r, "static/404.html")
			return
		}

		log.Println("get url error:", err)
		http.ServeFile(w, r, "static/500.html")
		return
	}

	if time.Now().After(expire) {
		http.ServeFile(w, r, "static/410.html")
		return
	}

	// Load into cache, already synchronized with DB
	h.cache.Put(hash, url, expire, false)

	// This request accessed it
	h.cache.Touch(hash)

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) CreateLongURL(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeResult(w, http.StatusBadRequest, resultErrorTmpl, struct{ Message string }{"Invalid request body."})
		return
	}
	ur := models.CreateShortURLRequest{URL: r.FormValue("url")}
	if !isValidURL(ur.URL) {
		writeResult(w, http.StatusBadRequest, resultErrorTmpl, struct{ Message string }{"URL must include http:// or https://"})
		return
	}
	tx, err := h.store.BeginTx()
	if err != nil {
		log.Println("start transaction fail:", err)
		writeResult(w, http.StatusInternalServerError, resultErrorTmpl, struct{ Message string }{"Something went wrong."})
		return
	}
	defer tx.Rollback()

	id, err := h.store.InsertURLToDB(tx, ur.URL)
	if err != nil {
		log.Println("insert error:", err)
		writeResult(w, http.StatusInternalServerError, resultErrorTmpl, struct{ Message string }{"Something went wrong."})
		return
	}
	hash := encode.EncodeBase62(id, h.MixMultiplierSecret)
	if err := h.store.UpdateHashToDB(tx, id, hash); err != nil {
		log.Println("update error:", err)
		writeResult(w, http.StatusInternalServerError, resultErrorTmpl, struct{ Message string }{"Something went wrong."})
		return
	}
	if err := tx.Commit(); err != nil {
		log.Println("commit error:", err)
		writeResult(w, http.StatusInternalServerError, resultErrorTmpl, struct{ Message string }{"Something went wrong."})
		return
	}

	shortURL := h.SiteURL + `/this/is/a/very/very/very/very/very/very/very/very/very/very/long/url/path/with/many/very/very/very/very/very/very/very/very/very/very/long/segments/that/keep/going/on/and/on/without/ending/until/it/becomes/an/extremely/long/url/path/with/more/very/very/very/very/very/very/very/very/very/very/very/very/long/parts/and/even/more/segments/that/continue/for/a/long/time/` + hash

	writeResult(w, http.StatusCreated, resultSuccessTmpl, struct{ ShortURL string }{shortURL})
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func (h *Handler) LongURLPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/long.html")
}

func (h *Handler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	data, err := os.ReadFile("static/404.html")
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	w.Write(data)
}
