package main

import(
	"log"
	"time"
	"encoding/json"
	"net/http"
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

func (h *Handler)CreateShortURL(w http.ResponseWriter,r *http.Request){
	var ur CreateShortURLRequest

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&ur)

	if err != nil{
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	id, err := h.store.InsertURLToDb(ur.URL)
	
	if err != nil {
		log.Println("insert error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return

	}
  
	hash := EncodeBase62(id)

	if err := h.store.UpdateHashToDb(id, hash); err != nil{
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	shortURL := h.config.SITE_URL + "/" + hash

	json.NewEncoder(w).Encode(map[string]string{"short_url" : shortURL})



}

func (h *Handler)RedirectURL(w http.ResponseWriter, r *http.Request){
	hash := r.PathValue("hash")

	time := time.Now()

	log.Println(hash)

	url , expire, err := GetURLFromDb(hash)

	if err != nil{
		log.Println("get url error:", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if expire


}
