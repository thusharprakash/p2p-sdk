package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"p2p-sdk/p2p"
)

func main() {
	nickFlag := flag.String("nick", "", "nickname to use in chat. will be generated if empty")
	roomFlag := flag.String("room", "example-room", "name of chat room to join")
	flag.Parse()

	ctx := context.Background()

	p2pInstance, err := p2p.NewP2P(ctx)
	if err != nil {
		panic(err)
	}

	nick := *nickFlag
	if len(nick) == 0 {
		nick = p2p.DefaultNick(p2pInstance.Host.ID())
	}

	roomName := *roomFlag

	room, err := p2pInstance.JoinRoom(ctx, roomName, nick)
	if err != nil {
		panic(err)
	}

	// Subscribe to message events
	p2pInstance.EventManager.Subscribe(p2p.EventTypeMessage, func(event p2p.EventMessage) {
		fmt.Printf("[%s] %s: %s\n", roomName, event.SenderNick, event.Data)
	})

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		if text == "/quit" {
			break
		}
		if err := room.Publish(p2p.EventTypeMessage, text); err != nil {
			fmt.Printf("Error publishing message: %s\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from stdin: %s\n", err)
	}

	select {}
}
