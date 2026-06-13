package reddit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store is a pure-Go SQLite store for parsed records and a crawl queue.
type Store struct {
	db *sql.DB
}

// OpenStore opens (creating if needed) the SQLite store at path and migrates it.
func OpenStore(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS records (
  entity_type TEXT NOT NULL,
  id          TEXT NOT NULL,
  url         TEXT,
  data        TEXT NOT NULL,
  fetched_at  TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (entity_type, id)
);
CREATE INDEX IF NOT EXISTS idx_records_type ON records(entity_type);

CREATE TABLE IF NOT EXISTS queue (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  url         TEXT NOT NULL UNIQUE,
  entity_type TEXT NOT NULL,
  priority    INTEGER NOT NULL DEFAULT 0,
  status      TEXT NOT NULL DEFAULT 'pending',
  html_path   TEXT,
  added_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_queue_status ON queue(status, priority DESC, id);
`
	_, err := s.db.Exec(schema)
	return err
}

// Put upserts a record keyed by (entityType, id).
func (s *Store) Put(entityType, id, url string, record any) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO records(entity_type,id,url,data) VALUES(?,?,?,?)
		 ON CONFLICT(entity_type,id) DO UPDATE SET url=excluded.url, data=excluded.data, fetched_at=datetime('now')`,
		entityType, id, url, string(data))
	return err
}

// Get reads a record's raw JSON by (entityType, id).
func (s *Store) Get(entityType, id string) ([]byte, error) {
	var data string
	err := s.db.QueryRow(`SELECT data FROM records WHERE entity_type=? AND id=?`, entityType, id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// Count returns the number of stored records of a type ("" for all types).
func (s *Store) Count(entityType string) (int, error) {
	var n int
	var err error
	if entityType == "" {
		err = s.db.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&n)
	} else {
		err = s.db.QueryRow(`SELECT COUNT(*) FROM records WHERE entity_type=?`, entityType).Scan(&n)
	}
	return n, err
}

// CountsByType returns a map of entity type to record count.
func (s *Store) CountsByType() (map[string]int, error) {
	rows, err := s.db.Query(`SELECT entity_type, COUNT(*) FROM records GROUP BY entity_type ORDER BY entity_type`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]int{}
	for rows.Next() {
		var t string
		var n int
		if err := rows.Scan(&t, &n); err != nil {
			return nil, err
		}
		out[t] = n
	}
	return out, rows.Err()
}

// Each streams stored records of a type to fn as raw JSON.
func (s *Store) Each(entityType string, fn func(id string, data []byte) error) error {
	var rows *sql.Rows
	var err error
	if entityType == "" {
		rows, err = s.db.Query(`SELECT id, data FROM records ORDER BY entity_type, id`)
	} else {
		rows, err = s.db.Query(`SELECT id, data FROM records WHERE entity_type=? ORDER BY id`, entityType)
	}
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id, data string
		if err := rows.Scan(&id, &data); err != nil {
			return err
		}
		if err := fn(id, []byte(data)); err != nil {
			return err
		}
	}
	return rows.Err()
}

// Enqueue adds a URL to the crawl queue (ignored if already present).
func (s *Store) Enqueue(ctx context.Context, url, entityType string, priority int) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO queue(url,entity_type,priority) VALUES(?,?,?) ON CONFLICT(url) DO NOTHING`,
		url, entityType, priority)
	return err
}

// NextPending claims and returns up to n pending queue items (status -> 'active').
func (s *Store) NextPending(ctx context.Context, n int) ([]QueueItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx,
		`SELECT id,url,entity_type,priority FROM queue WHERE status='pending'
		 ORDER BY priority DESC, id LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	var items []QueueItem
	for rows.Next() {
		var it QueueItem
		if err := rows.Scan(&it.ID, &it.URL, &it.EntityType, &it.Priority); err != nil {
			_ = rows.Close()
			return nil, err
		}
		items = append(items, it)
	}
	_ = rows.Close()
	for _, it := range items {
		if _, err := tx.ExecContext(ctx, `UPDATE queue SET status='active' WHERE id=?`, it.ID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return items, nil
}

// MarkFetched records a queue item as fetched with its cache path.
func (s *Store) MarkFetched(ctx context.Context, id int64, htmlPath string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE queue SET status='fetched', html_path=? WHERE id=?`, htmlPath, id)
	return err
}

// MarkFailed records a queue item as failed.
func (s *Store) MarkFailed(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE queue SET status='failed' WHERE id=?`, id)
	return err
}

// QueueStats returns counts of queue items grouped by status.
func (s *Store) QueueStats() (map[string]int, error) {
	rows, err := s.db.Query(`SELECT status, COUNT(*) FROM queue GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]int{}
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			return nil, err
		}
		out[st] = n
	}
	return out, rows.Err()
}

// ResetActive returns any 'active' items to 'pending' (recover from a crash).
func (s *Store) ResetActive() error {
	_, err := s.db.Exec(`UPDATE queue SET status='pending' WHERE status='active'`)
	return err
}

// Vacuum compacts the database file.
func (s *Store) Vacuum() error {
	_, err := s.db.Exec(`VACUUM`)
	return err
}

// Info returns a human summary of the store's contents.
func (s *Store) Info() (string, error) {
	counts, err := s.CountsByType()
	if err != nil {
		return "", err
	}
	q, err := s.QueueStats()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("records=%v queue=%v", counts, q), nil
}
