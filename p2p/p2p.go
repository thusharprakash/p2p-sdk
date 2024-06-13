package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
)

const (
	DiscoveryServiceTag = "p2p-discovery"
	EventRoomBufSize    = 128
)

type PeerToPeer struct {
	Host         host.Host
	PubSub       *pubsub.PubSub
	Rooms        map[string]*EventRoom
	EventManager *EventManager
	Storage      *Storage
	id           string
}

type EventMessage struct {
	EventType   string
	Data        string
	SenderID    string
	SenderNick  string
	Timestamp   int64
	VectorClock VectorClock
}

func NewP2P(ctx context.Context) (*PeerToPeer, error) {
	fmt.Println("Initializing P2P")
	h, err := libp2p.New(
		libp2p.Security(noise.ID, noise.New),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		return nil, err
	}

	fmt.Println("Initializing PubSub")
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	fmt.Println("Setting up discovery")
	if err := SetupDiscovery(h); err != nil {
		return nil, err
	}

	em := NewEventManager()

	p2p := &PeerToPeer{
		Host:         h,
		PubSub:       ps,
		Rooms:        make(map[string]*EventRoom),
		EventManager: em,
		id:           h.ID().ShortString(),
	}
	return p2p, nil
}

func (p2p *PeerToPeer) SetEventStorage(storage *Storage) {
	p2p.Storage = storage

	// Load events from storage and dispatch them
	events, err := storage.GetEvents()
	if err != nil {
		fmt.Printf("Error retrieving events from storage: %s\n", err)
		return
	}

	for _, event := range events {
		p2p.EventManager.DispatchWithOrdering(event)
	}
}

func (p2p *PeerToPeer) JoinRoom(ctx context.Context, roomName, nick string) (*EventRoom, error) {
	topic, err := p2p.PubSub.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	room := &EventRoom{
		ctx:         ctx,
		ps:          p2p.PubSub,
		topic:       topic,
		sub:         sub,
		self:        p2p.Host,
		nick:        nick,
		roomName:    roomName,
		Messages:    make(chan *EventMessage, EventRoomBufSize),
		VectorClock: make(VectorClock),
		Storage:     p2p.Storage,
	}

	p2p.Rooms[roomName] = room

	// send an inital message to the room with all the events
	go room.SendEventsToPeer(p2p.Host.ID())

	go room.readLoop(p2p.EventManager)
	return room, nil
}
