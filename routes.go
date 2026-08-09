package veryshort

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /shorten", h.CreateShortURL)
	mux.HandleFunc("GET /{hash}", h.RedirectURL)
	mux.HandleFunc("GET /{$}", h.HomePage)
	mux.HandleFunc("POST /makeitlong", h.CreateLongURL)
	mux.HandleFunc("GET /this/is/a/very/very/very/very/very/very/very/very/very/very/long/url/path/with/many/very/very/very/very/very/very/very/very/very/very/long/segments/that/keep/going/on/and/on/without/ending/until/it/becomes/an/extremely/long/url/path/with/more/very/very/very/very/very/very/very/very/very/very/very/very/long/parts/and/even/more/segments/that/continue/for/a/long/time/{hash}", h.RedirectURL)
	mux.HandleFunc("GET /long", h.LongURLPage)
	mux.Handle("GET /static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static")),
		),
	)
	mux.HandleFunc("GET /", h.NotFoundHandler)
}
