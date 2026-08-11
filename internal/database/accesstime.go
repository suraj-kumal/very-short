package database

import (
	"github.com/suraj-kumal/very-short/internal/contracts"
)

func (s *Store) UpdateLastAccessTimes(
	nodes []contracts.DirtyNode,
) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE url_data
		SET LastAccessTime = ?
		WHERE hash = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, node := range nodes {
		if _, err := stmt.Exec(
			node.LastAccessTime,
			node.Hash,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) GetHotURLs(limit int) ([]URLRecord, error) {
	rows, err := s.conn.Query(`
		SELECT hash, url, expire_at, LastAccessTime
		FROM url_data
		WHERE expire_at > NOW()
		ORDER BY LastAccessTime DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []URLRecord

	for rows.Next() {
		var record URLRecord

		if err := rows.Scan(
			&record.Hash,
			&record.URL,
			&record.Expire,
			&record.LastAccessTime,
		); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, rows.Err()
}
