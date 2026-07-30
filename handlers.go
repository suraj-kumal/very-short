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
    InsertURLToDb(url string) (int, error)
    UpdateHashToDb(id int, hash string) error
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

	if !isValidURL(ur.URL) {
		http.Error(w, "invalid url: must include http:// or https://", http.StatusBadRequest)
		return
	}	

	if err != nil{
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	id, err := h.store.InsertURLToDb(ur.URL)
	
	if err != nil {
		log.Println("insert error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return

	}
  
	hash := EncodeBase62(id, h.config.MixMultiplierSecret)

	if err := h.store.UpdateHashToDb(id, hash); err != nil{
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
