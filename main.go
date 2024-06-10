package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"p2p-sdk/p2p"
)

func main() {
	nickFlag := flag.String("nick", "", "nickname to use in chat. will be generated if empty")
	roomFlag := flag.String("room", "example-room", "name of chat room to join")
	storagePathFlag := flag.String("storage", "events.db", "path to SQLite database file")
	flag.Parse()

	ctx := context.Background()

	// Initialize the storage
	storage, err := p2p.NewStorage(*storagePathFlag)
	if err != nil {
		panic(err)
	}

	// Initialize the P2P library
	p2pInstance, err := p2p.NewP2P(ctx)
	if err != nil {
		panic(err)
	}

	// Set the storage for P2P instance
	p2pInstance.SetEventStorage(storage)

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
		// Save event to storage
		if err := storage.SaveEvent(event); err != nil {
			fmt.Printf("Error saving event: %s\n", err)
		}
	})

	// Create a set of initial events for testing
	for i := 1; i <= 5; i++ {
		eventData := fmt.Sprintf("Initial event %d", i)
		if err := room.Publish(p2p.EventTypeMessage, eventData); err != nil {
			fmt.Printf("Error publishing initial event: %s\n", err)
		}
		time.Sleep(1 * time.Second)
	}

	// Retrieve and print events from storage to verify persistence
	events, err := storage.GetEvents()
	if err != nil {
		fmt.Printf("Error retrieving events: %s\n", err)
	} else {
		fmt.Println("Retrieved events from storage:")
		for _, event := range events {
			fmt.Printf("[%s] %s: %s\n", roomName, event.SenderNick, event.Data)
		}
	}

	// Start interactive mode for user input
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
