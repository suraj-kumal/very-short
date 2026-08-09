package main

import (
	"log"
	"net/http"
	"time"

	veryshort "github.com/suraj-kumal/very-short"
)

func main() {
	config := veryshort.Load()

	conn, err := veryshort.DatabaseConnection(config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.Ping(); err != nil {
		log.Fatal("database unreachable:", err)
	}
	defer conn.Close()

	store := veryshort.NewStore(conn)
	cache := veryshort.NewCache(10000)

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

	timeSync := veryshort.NewSyncState()
	timeSync.Set(time.Now())

	h := veryshort.New(store, config, cache, timeSync)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	log.Println("listening on:", config.PORT)
	if err := http.ListenAndServe(config.PORT, mux); err != nil {
		log.Println("server error:", err)
	}
}
