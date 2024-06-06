package p2p

import (
	"context"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	DiscoveryServiceTag = "p2p-discovery"
	EventRoomBufSize    = 128
)

type P2P struct {
	Host   host.Host
	PubSub *pubsub.PubSub
	Rooms  map[string]*EventRoom
}

type EventRoom struct {
	Messages chan *EventMessage

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string
}

type EventMessage struct {
	Message    string
	SenderID   string
	SenderNick string
}

func NewP2P(ctx context.Context) (*P2P, error) {
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	if err := SetupDiscovery(h); err != nil {
		return nil, err
	}

	p2p := &P2P{
		Host:   h,
		PubSub: ps,
		Rooms:  make(map[string]*EventRoom),
	}
	return p2p, nil
}

func (p2p *P2P) JoinRoom(ctx context.Context, roomName, nick string) (*EventRoom, error) {
	topic, err := p2p.PubSub.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	room := &EventRoom{
		ctx:      ctx,
		ps:       p2p.PubSub,
		topic:    topic,
		sub:      sub,
		self:     p2p.Host.ID(),
		nick:     nick,
		roomName: roomName,
		Messages: make(chan *EventMessage, EventRoomBufSize),
	}

	p2p.Rooms[roomName] = room

	go room.readLoop()
	return room, nil
}
