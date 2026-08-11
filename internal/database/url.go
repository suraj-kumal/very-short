package database

import (
	"database/sql"
	"time"
)

func (s *Store) InsertURLToDB(tx *sql.Tx, url string) (int, error) {
	res, err := tx.Exec(
		"INSERT INTO url_data (url) VALUES (?)",
		url,
	)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastID), nil
}

func (s *Store) UpdateHashToDB(tx *sql.Tx, id int, hash string) error {
	_, err := tx.Exec(
		"UPDATE url_data SET hash = ? WHERE id = ?",
		hash,
		id,
	)

	return err
}

func (s *Store) GetURLFromDB(
	hash string,
) (string, time.Time, error) {
	var url string
	var expire time.Time

	err := s.conn.QueryRow(
		"SELECT url, expire_at FROM url_data WHERE hash = ?",
		hash,
	).Scan(&url, &expire)
	if err != nil {
		return "", time.Time{}, err
	}

	return url, expire, nil
}
