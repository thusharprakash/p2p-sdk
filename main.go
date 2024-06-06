// Description: This is a simple chat application that uses the P2P library to create a chat room and send messages to other peers in the room.

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
	// parse some flags to set our nickname and the room to join
	nickFlag := flag.String("nick", "", "nickname to use in chat. will be generated if empty")
	roomFlag := flag.String("room", "example-room", "name of chat room to join")
	flag.Parse()

	ctx := context.Background()

	// Initialize the P2P library
	p2pInstance, err := p2p.NewP2P(ctx)
	if err != nil {
		panic(err)
	}

	// use the nickname from the cli flag, or a default if blank
	nick := *nickFlag
	if len(nick) == 0 {
		nick = p2p.DefaultNick(p2pInstance.Host.ID())
	}

	// join the room from the cli flag, or the flag default
	roomName := *roomFlag

	// Join an event room
	room, err := p2pInstance.JoinRoom(ctx, roomName, nick)
	if err != nil {
		panic(err)
	}

	// Listen for incoming messages
	go func() {
		for msg := range room.Messages {
			fmt.Printf("[%s] %s: %s\n", roomName, msg.SenderNick, msg.Message)
		}
	}()

	// Read user input from stdin and publish to the room
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		if text == "/quit" {
			break
		}
		if err := room.Publish(text); err != nil {
			fmt.Printf("Error publishing message: %s\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from stdin: %s\n", err)
	}

	// Keep the main function running
	select {}
}
