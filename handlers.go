package main

import(
	"log"
	"time"
	"encoding/json"
	"net/http"
	"errors"
	"database/sql"
	"net/url"
)

type UrlStore interface {
		BeginTx()(*sql.Tx, error)

    InsertURLToDb(tx *sql.Tx,url string) (int, error)
		UpdateHashToDb(tx *sql.Tx,id int, hash string) error
		GetURLFromDb(hash string) (string , time.Time, error)

}

type Handler struct{
	store UrlStore
	config Config
}

func New(s UrlStore, config Config) *Handler{
	return &Handler{
		store: s,
		config: config,

}
}



func isValidURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func (h *Handler)CreateShortURL(w http.ResponseWriter,r *http.Request){


	var ur CreateShortURLRequest



	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&ur)

	if err != nil{
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
	id, err := h.store.InsertURLToDb(tx,ur.URL)
	log.Println("insert took:", time.Since(t2))

	
	if err != nil {
		log.Println("insert error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return

	}

  
	hash := EncodeBase62(id, h.config.MixMultiplierSecret)
	
	t3 := time.Now()

	if err := h.store.UpdateHashToDb(tx, id, hash); err != nil{
		log.Println("update error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	log.Println("update took:", time.Since(t3))

	t4 := time.Now()

	err = tx.Commit()
	log.Println("commit took:", time.Since(t4))

	if err != nil{
			log.Println("update error:", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return

	}


	shortURL := h.config.SITE_URL + "/" + hash


	json.NewEncoder(w).Encode(map[string]string{"short_url" : shortURL})



}


func (h *Handler)RedirectURL(w http.ResponseWriter, r *http.Request){
	hash := r.PathValue("hash")

	now := time.Now()

	log.Println(hash)

	url , expire, err := h.store.GetURLFromDb(hash)

	if err != nil{

		if errors.Is(err, sql.ErrNoRows){
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		log.Println("get url error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if now.After(expire){
		http.Error(w, "url has expire", http.StatusGone)
		return
	}

	http.Redirect(w,r,url, http.StatusMovedPermanently)

}
