package p2p

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db   *sql.DB
	mu   sync.Mutex
	path string
}

func NewStorage(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		db:   db,
		path: path,
	}

	if err := storage.init(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) init() error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT,
		data TEXT,
		sender_id TEXT,
		sender_nick TEXT,
		timestamp INTEGER,
		vector_clock TEXT
	);`

	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

func (s *Storage) SaveEvent(event EventMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vectorClockJSON, err := json.Marshal(event.VectorClock)
	if err != nil {
		return fmt.Errorf("failed to marshal vector clock: %v", err)
	}

	insertSQL := `INSERT INTO events (event_type, data, sender_id, sender_nick, timestamp, vector_clock) VALUES (?, ?, ?, ?, ?, ?);`
	_, err = s.db.Exec(insertSQL, event.EventType, event.Data, event.SenderID, event.SenderNick, event.Timestamp, string(vectorClockJSON))
	if err != nil {
		return fmt.Errorf("failed to insert event: %v", err)
	}

	return nil
}

func (s *Storage) GetEvents() ([]EventMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query("SELECT event_type, data, sender_id, sender_nick, timestamp, vector_clock FROM events")
	if err != nil {
		return nil, fmt.Errorf("failed to select events: %v", err)
	}
	defer rows.Close()

	var events []EventMessage
	for rows.Next() {
		var event EventMessage
		var vectorClockJSON string

		err = rows.Scan(&event.EventType, &event.Data, &event.SenderID, &event.SenderNick, &event.Timestamp, &vectorClockJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %v", err)
		}

		err = json.Unmarshal([]byte(vectorClockJSON), &event.VectorClock)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vector clock: %v", err)
		}

		events = append(events, event)
	}

	return events, nil
} 