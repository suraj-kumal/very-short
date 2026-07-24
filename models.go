package very_short


import "time"

type url_data struct{
	id int
	url string
	hash string
	created_at time.Time
	expire_at time.Time
}
