package main

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /shorten", h.CreateShortURL)
	mux.HandleFunc("GET /{hash}", h.RedirectURL)
	mux.HandleFunc("GET /", h.HomePage)
	mux.Handle("GET /static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static")),
		),
	)
}
