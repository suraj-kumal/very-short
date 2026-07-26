package very_short

import(
	"encoding/json"
	"net/http"
	"strconv"
)

type UrlStore interface{
	CreateShortURL(ud url_data)(url_data, error)
}

type Handler struct{
	store UrlStore
}

func New(s UrlStore) *Handler{
	return &Handler{store: s}
}

func (h *Handler)CreateShortURL(w http.ResponseWriter,r *http.Request){
	var ur CreateShortURLRequest

	decoder := json.NewDecoder(r.body)

	err := decoder.Decode(&ur)

	if err != nil{
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}




}
