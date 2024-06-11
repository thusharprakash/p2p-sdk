package p2p

type PeerMessageCallback interface {
	OnMessage(string)
}