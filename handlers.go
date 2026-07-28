package main

import(
	"log"
	"encoding/json"
	"net/http"
)

type UrlStore interface {
    InsertURLToDb(url string) (int, error)
    UpdateHashToDb(id int, hash string) error
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
		http.Error(w, "something went wrong while insering", http.StatusInternalServerError)
		return

	}
  
	hash := EncodeBase62(id)

	if err := h.store.UpdateHashToDb(id, hash); err != nil{
		http.Error(w, "something went wrong updating", http.StatusInternalServerError)
		return
	}

	shortURL := h.config.SITE_URL + "/" + hash

	json.NewEncoder(w).Encode(map[string]string{"short_url" : shortURL})



}

func (h *Handler)Redirect(w http.ResponseWriter, r *http.Request){

}
