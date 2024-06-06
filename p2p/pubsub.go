package p2p

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/peer"
)

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
