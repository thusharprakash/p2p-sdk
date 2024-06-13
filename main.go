package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"p2p-sdk/p2p"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize P2P
	p2pNode, err := p2p.NewP2P(ctx)
	if err != nil {
		fmt.Printf("Error initializing P2P: %v\n", err)
		return
	}

	// Generate a random database name and nickname
	dbName := fmt.Sprintf("events_%d.db", rand.Int())
	nickName := fmt.Sprintf("node%d", rand.Intn(1000))

	// Ensure the data directory exists
	dataDir := "data"
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		return
	}

	// Initialize storage
	storage, err := p2p.NewStorage(filepath.Join(dataDir, dbName))
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		return
	}

	p2pNode.SetEventStorage(storage)

	// Initialize Event Manager
	eventManager := p2p.NewEventManager()

	// Register event handler
	eventManager.RegisterEventHandler("message", func(event p2p.EventMessage) {
		fmt.Printf("Received event from %s: %s\n", event.SenderNick, event.Data)
		if err := storage.AddEventIfNotDuplicate(event); err != nil {
			fmt.Printf("Error adding event to storage: %v\n", err)
		}
	})

	// Sync with existing peers
	existingEvents, err := storage.GetEvents()
	if err != nil {
		fmt.Printf("Error retrieving existing events: %v\n", err)
		return
	}

	for _, event := range existingEvents {
		eventManager.DispatchWithOrdering(event)
	}

	// Join room
	room, err := p2pNode.JoinRoom(ctx, "test-room", nickName)
	if err != nil {
		fmt.Printf("Error joining room: %v\n", err)
		return
	}

	// Periodic synchronization
	// go storage.PeriodicSync(eventManager, p2pNode.Host.Peerstore().Peers(), p2pNode.Host, 30*time.Second)
	go storage.PeriodicSync(ctx, room, 30*time.Second)

	// Publish random events
	go func() {
		for {
			time.Sleep(5 * time.Second)
			data := fmt.Sprintf("Random Event: %d", rand.Int())
			if err := room.Publish("message", data); err != nil {
				fmt.Printf("Error publishing random event: %v\n", err)
			}
		}
	}()

	// Send a chat message to all devices
	go func() {
		time.Sleep(10 * time.Second)
		data := "Hello, this is a chat message to all devices!"
		if err := room.Publish("message", data); err != nil {
			fmt.Printf("Error publishing chat message: %v\n", err)
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("Shutting down...")
}
