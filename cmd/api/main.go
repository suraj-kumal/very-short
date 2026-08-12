package main

import (
	"log"
	"net/http"
	"time"

	"github.com/suraj-kumal/very-short/internal/cache"
	"github.com/suraj-kumal/very-short/internal/config"
	"github.com/suraj-kumal/very-short/internal/database"
	"github.com/suraj-kumal/very-short/internal/handlers"
	"github.com/suraj-kumal/very-short/internal/timesync"
)

func main() {
	cfg := config.Load()

	conn, err := database.DatabaseConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.Ping(); err != nil {
		log.Fatal("database unreachable:", err)
	}
	defer conn.Close()

	store := database.DatabaseStore(conn)
	cache := cache.NewCache(10000)

	// Cache warming — DB order is hottest-first (last_access_time DESC).
	// Put() does addToFront, so insert in reverse to preserve that as MRU order.
	records, err := store.GetHotURLs(1000)
	if err != nil {
		log.Fatal("cache warming failed:", err)
	}
	for i := len(records) - 1; i >= 0; i-- {
		r := records[i]
		cache.Put(r.Hash, r.URL, r.Expire, false)
	}

	timeSync := timesync.NewSyncState()
	timeSync.Set(time.Now())

	h := handlers.HandlerStore(store, cfg.SiteURL, cfg.MixMultiplierSecret, cache, timeSync)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	log.Println("listening on:", cfg.PORT)
	if err := http.ListenAndServe(cfg.PORT, mux); err != nil {
		log.Println("server error:", err)
	}
}
