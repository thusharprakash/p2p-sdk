package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db   *bolt.DB
	mu   sync.Mutex
	path string
}

func NewStorage(path string) (*Storage, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db, path: path}, nil
}

func (s *Storage) AddEvent(event EventMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("events"))
		if err != nil {
			return err
		}
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		return b.Put([]byte(event.ID()), data)
	})
}

func (s *Storage) AddEventIfNotDuplicate(event EventMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("events"))
		if err != nil {
			return err
		}
		if b.Get([]byte(event.ID())) != nil {
			return nil // Duplicate event, do not store
		}
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		return b.Put([]byte(event.ID()), data)
	})
}

func (s *Storage) GetEvents() ([]EventMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var events []EventMessage
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("events"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var event EventMessage
			if err := json.Unmarshal(v, &event); err != nil {
				return err
			}
			events = append(events, event)
			return nil
		})
	})
	return events, err
}

func (event *EventMessage) ID() string {
	return fmt.Sprintf("%s:%d", event.SenderID, event.Timestamp)
}

func (s *Storage) PeriodicSync(em *EventManager, peers []peer.ID, host host.Host, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		events, err := s.GetEvents()
		if err != nil {
			fmt.Printf("Error retrieving events: %v\n", err)
			continue
		}
		for _, event := range events {
			em.DispatchWithOrdering(event)
			em.GossipEvent(context.Background(), event, peers, host)
		}
	}
}
