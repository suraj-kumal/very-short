package main

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// conn data type and it points to sql.DB
type Store struct {
	conn *sql.DB
}

func DatabaseConnection(databaseURL string) (*sql.DB, error) {
	return sql.Open("mysql", databaseURL)
}

// recieves database dependency and injects it into the store
func NewStore(conn *sql.DB) *Store {
	return &Store{conn: conn}
}

func (s *Store) InsertURLToDb(url string) (id int, err error) {
	res, err := s.conn.Exec("INSERT INTO url_data (url) VALUES (?)", url)

	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(lastID), nil
}

func (s *Store) UpdateHashToDb(id int, hash string) error {
	_, err := s.conn.Exec("UPDATE url_data SET hash = ? WHERE id = ?", hash, id)

	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetURLFromDb(hash string)(url string , expire time.Time, err error){

	var url string
	var expire time.Time

	err = s.conn.QueryRow("SELECT url, expire_at FROM url_data WHERE hash = ?", hash).Scan(&url, &expire)

	if err != nil{
		return "", time.Time{}, err
		
	}

	return url, expire, nil


}
