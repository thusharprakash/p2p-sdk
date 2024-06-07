package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	DiscoveryServiceTag = "p2p-discovery"
	EventRoomBufSize    = 128
)


func NewP2P(ctx context.Context) (*PeerToPeer, error) {
	fmt.Println("Initializing P2P")
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
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

	fmt.Println("Done returning")
	p2p := &PeerToPeer{
		Host:   h,
		PubSub: ps,
		Rooms:  make(map[string]*EventRoom),
	}
	return p2p, nil
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
