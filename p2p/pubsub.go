package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type EventRoom struct {
	ctx         context.Context
	ps          *pubsub.PubSub
	topic       *pubsub.Topic
	sub         *pubsub.Subscription
	self        host.Host
	nick        string
	roomName    string
	Messages    chan *EventMessage
	VectorClock VectorClock
	Storage     *Storage
	P2P         *PeerToPeer
}

func (room *EventRoom) Publish(eventType, data string) error {
	room.VectorClock.Increment(room.self.ID().String())
	m := EventMessage{
		EventType:   eventType,
		Data:        data,
		SenderID:    room.self.ID().String(),
		SenderNick:  room.nick,
		Timestamp:   time.Now().Unix(),
		VectorClock: room.VectorClock.Copy(),
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

func topicName(roomName string) string {
	return "event-room:" + roomName
}

func (room *EventRoom) readLoop(em *EventManager) {
	for {
		msg, err := room.sub.Next(room.ctx)
		if err != nil {
			close(room.Messages)
			return
		}
		// if msg.ReceivedFrom == room.self.ID() {
		// 	continue
		// }
		evt := new(EventMessage)
		err = json.Unmarshal(msg.Data, evt)
		if err != nil {
			continue
		}
		room.VectorClock.Update(evt.VectorClock)
		room.Messages <- evt
		em.DispatchWithOrdering(*evt)

		// Save event to storage
		if room.Storage != nil {
			if err := room.Storage.AddEventIfNotDuplicate(*evt); err != nil {
				fmt.Printf("Error saving event: %s\n", err)
			}
		}
	}
}
