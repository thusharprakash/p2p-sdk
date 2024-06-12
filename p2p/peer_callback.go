package p2p

type PeerCallback interface {
	OnMessage(string)
}