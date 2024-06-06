package p2p

import (
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
)

func printErr(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}

func DefaultNick(p peer.ID) string {
	return fmt.Sprintf("%s-%s", os.Getenv("USER"), shortID(p))
}

func shortID(p peer.ID) string {
	pretty := p.String()
	return pretty[len(pretty)-8:]
}
