package very_short

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux){
	mux.HandleFunc("POST /shorten", h.CreateShortURL)
}
