package p2p

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
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
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY,
		event_type TEXT,
		data TEXT,
		sender_id TEXT,
		sender_nick TEXT,
		timestamp INTEGER,
		vector_clock TEXT
	)`)
	return err
}

func (s *Storage) AddEvent(event EventMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`INSERT INTO events (event_type, data, sender_id, sender_nick, timestamp, vector_clock) VALUES (?, ?, ?, ?, ?, ?)`,
		event.EventType, event.Data, event.SenderID, event.SenderNick, event.Timestamp, event.VectorClock.String())
	return err
}

func (s *Storage) GetEvents() ([]EventMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT event_type, data, sender_id, sender_nick, timestamp, vector_clock FROM events`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventMessage
	for rows.Next() {
		var event EventMessage
		var vc string
		if err := rows.Scan(&event.EventType, &event.Data, &event.SenderID, &event.SenderNick, &event.Timestamp, &vc); err != nil {
			return nil, err
		}
		event.VectorClock, err = ParseVectorClock(vc)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) AddEventIfNotDuplicate(event EventMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM events WHERE event_type = ? AND data = ? AND sender_id = ? AND timestamp = ? AND vector_clock = ?`,
		event.EventType, event.Data, event.SenderID, event.Timestamp, event.VectorClock.String()).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Duplicate event found, discard it
		return nil
	}

	return s.AddEvent(event)
}

func (s *Storage) SyncEvents(events []EventMessage) error {
	for _, event := range events {
		if err := s.AddEventIfNotDuplicate(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) PeriodicSync(eventManager *EventManager, peers []peer.ID, host host.Host, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		events, err := s.GetEvents()
		if err != nil {
			fmt.Printf("Error retrieving events: %v\n", err)
			continue
		}

		for _, event := range events {
			eventManager.DispatchWithOrdering(event)
			eventManager.GossipEvent(context.Background(), event, peers, host)
		}
	}
}
