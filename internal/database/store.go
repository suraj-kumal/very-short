package database

import "database/sql"

type Store struct {
	conn *sql.DB
}

func DatabaseStore(conn *sql.DB) *Store {
	return &Store{conn: conn}
}

func (s *Store) BeginTx() (*sql.Tx, error) {
	return s.conn.Begin()
}
