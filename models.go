package main

import "time"

type URL_data struct{
	id int
	url string
	hash string
	created_at time.Time
	expire_at time.Time
}

type CreateShortURLRequest struct{
	URL string `json:"url"`
}
