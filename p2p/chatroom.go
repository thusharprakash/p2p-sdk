package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const (
	// DiscoveryInterval is how often we re-publish our mDNS records.
	DiscoveryInterval = time.Hour

	// DiscoveryServiceTag is used in our mDNS advertisements to discover other peers.
	DiscoveryServiceTag = "event-lib"

	// EventRoomBufSize is the number of incoming messages to buffer for each topic.
	EventRoomBufSize = 128
)

// EventLib is the main struct for the event passing library
type EventLib struct {
	Host   host.Host
	PubSub *pubsub.PubSub
	Rooms  map[string]*EventRoom
}

// EventRoom represents a subscription to a single PubSub topic.
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

// EventMessage gets converted to/from JSON and sent in the body of pubsub messages.
type EventMessage struct {
	Message    string
	SenderID   string
	SenderNick string
}

// NewEventLib initializes a new EventLib instance
func NewEventLib(ctx context.Context) (*EventLib, error) {
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	if err := setupDiscovery(h); err != nil {
		return nil, err
	}

	el := &EventLib{
		Host:   h,
		PubSub: ps,
		Rooms:  make(map[string]*EventRoom),
	}
	return el, nil
}

// JoinRoom joins a pubsub room
func (el *EventLib) JoinRoom(ctx context.Context, roomName, nick string) (*EventRoom, error) {
	topic, err := el.PubSub.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	room := &EventRoom{
		ctx:      ctx,
		ps:       el.PubSub,
		topic:    topic,
		sub:      sub,
		self:     el.Host.ID(),
		nick:     nick,
		roomName: roomName,
		Messages: make(chan *EventMessage, EventRoomBufSize),
	}

	el.Rooms[roomName] = room

	go room.readLoop()
	return room, nil
}

// Publish sends a message to the pubsub topic.
func (room *EventRoom) Publish(message string) error {
	m := EventMessage{
		Message:    message,
		SenderID:   room.self.String(),
		SenderNick: room.nick,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return room.topic.Publish(room.ctx, msgBytes)
}

func (room *EventRoom) ListPeers() []peer.ID {
	return room.ps.ListPeers(topicName(room.roomName))
}

func (room *EventRoom) readLoop() {
	for {
		msg, err := room.sub.Next(room.ctx)
		if err != nil {
			close(room.Messages)
			return
		}
		if msg.ReceivedFrom == room.self {
			continue
		}
		em := new(EventMessage)
		err = json.Unmarshal(msg.Data, em)
		if err != nil {
			continue
		}
		room.Messages <- em
	}
}

func topicName(roomName string) string {
	return "event-room:" + roomName
}

func setupDiscovery(h host.Host) error {
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID)
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
}

func printErr(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}

func defaultNick(p peer.ID) string {
	return fmt.Sprintf("%s-%s", os.Getenv("USER"), shortID(p))
}

func shortID(p peer.ID) string {
	pretty := p.String()
	return pretty[len(pretty)-8:]
}
