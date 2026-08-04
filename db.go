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

func (s *Store) BeginTx() (*sql.Tx, error) {
	return s.conn.Begin()
}

func (s *Store) InsertURLToDB(tx *sql.Tx, url string) (id int, err error) {
	res, err := tx.Exec("INSERT INTO url_data (url) VALUES (?)", url)
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
	_, err := tx.Exec("UPDATE url_data SET hash = ? WHERE id = ?", hash, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetURLFromDB(hash string) (url string, expire time.Time, err error) {
	err = s.conn.QueryRow("SELECT url, expire_at FROM url_data WHERE hash = ?", hash).Scan(&url, &expire)
	if err != nil {
		return "", time.Time{}, err
	}

	return url, expire, nil
}

func (s *Store) UpdateLastAccessTimes(nodes []DirtyNode) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE url_data
		SET last_access_time = ?
		WHERE hash = ?
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, node := range nodes {
		_, err := stmt.Exec(
			node.LastAccessTime,
			node.Hash,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

type URLRecord struct {
	Hash           string
	URL            string
	Expire         time.Time
	LastAccessTime time.Time
}

func (s *Store) GetHotURLs(limit int) ([]URLRecord, error) {
	rows, err := s.conn.Query(`
		SELECT hash, url, expire_at, last_access_time
		FROM url_data
		WHERE expire_at > NOW()
		ORDER BY last_access_time DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var records []URLRecord

	for rows.Next() {
		var record URLRecord

		err := rows.Scan(
			&record.Hash,
			&record.URL,
			&record.Expire,
			&record.LastAccessTime,
		)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, rows.Err()
}
