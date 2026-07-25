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
