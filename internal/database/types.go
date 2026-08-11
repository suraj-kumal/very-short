package database

import "time"

type URLRecord struct {
	Hash           string
	URL            string
	Expire         time.Time
	LastAccessTime time.Time
}
