package p2p

import (
	"context"
	"fmt"
	"time"

	internal "p2p-sdk/p2p/internals"

	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/zap"
)

func StartP2PChat(config *NodeConfig){
	nickFlag := GenerateRandomString(6)
	roomFlag := "test-chat-room-dabzee"

	if config == nil {
		config = NewNodeConfig()
	}

	// Set up netdriver.
	if config.netDriver != nil {
		logger, _ := zap.NewDevelopment()
		inet := &inet{
			net:    config.netDriver,
			logger: logger,
		}
		
		internal.SetNetDriver(inet)
		manet.SetNetInterface(inet)
	}


	mdnsLocked := false

	if config.mdnsLockerDriver != nil {
		config.mdnsLockerDriver.Lock()
		mdnsLocked = true
	}
	ctx := context.Background()

	// Initialize the P2P library
	p2pInstance, err := NewP2P(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("P2P instance created with ID %s\n", p2pInstance.Host.ID())
	// use the nickname from the cli flag, or a default if blank
	nick := nickFlag
	if len(nickFlag) == 0 {
		nick = DefaultNick(p2pInstance.Host.ID())
	}

	// join the room from the cli flag, or the flag default
	roomName := roomFlag

	// Join an event room
	room, err := p2pInstance.JoinRoom(ctx, roomName, nick)

	fmt.Printf("Joined room %s as %s\n", roomName, nick)
	if err != nil {
		panic(err)
	}

	if mdnsLocked && config.mdnsLockerDriver != nil{
		config.mdnsLockerDriver.Unlock()
	}

	// Listen for incoming messages
	go func() {
		for msg := range room.Messages {
			fmt.Printf("[%s] %s: %s\n", roomName, msg.SenderNick, msg.Message)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

	for {
        select {
        case <-ticker.C:
            currentTime := time.Now().Format("2006-01-02 15:04:05")
			if err := room.Publish(currentTime); err != nil {
				fmt.Printf("Error publishing message: %s\n", err)
			}
        }
    }
}