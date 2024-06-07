package p2p

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

type PeerToPeer struct {
	Host   host.Host
	PubSub *pubsub.PubSub
	Rooms  map[string]*EventRoom
}