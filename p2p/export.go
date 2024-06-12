package p2p

import (
	"context"
	"fmt"
	"time"

	internal "p2p-sdk/p2p/internals"

	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/zap"
)

var storage *Storage
var p2pNode *PeerToPeer
var globalRoom *EventRoom

type PeerMessageData struct{
	message string
	sender string
}

func StartP2PChat(config *NodeConfig)(string){

	if config == nil {
		config = NewNodeConfig()
		config.SetNickName(GenerateRandomString(10))
	}
	nickFlag := config.nickName
	roomFlag := "test-chat-room-dabzee"

	// Initialize the storage
	newStorage,err := NewStorage(config.storagePath)
	storage = newStorage
	if err != nil {
		panic(err)
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
	p2pNode = p2pInstance
	if err != nil {
		panic(err)
	}

	p2pInstance.SetEventStorage(storage)

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

	globalRoom = room
	fmt.Printf("Joined room %s as %s\n", roomName, nick)
	if err != nil {
		panic(err)
	}

	if mdnsLocked && config.mdnsLockerDriver != nil{
		config.mdnsLockerDriver.Unlock()
	}

	// Sync with existing peers
	existingEvents, err := storage.GetEvents()
	if err != nil {
		fmt.Printf("Error retrieving existing events: %v\n", err)
	}
	fmt.Println("Existing events",existingEvents)
	for _, event := range existingEvents {
		p2pInstance.EventManager.DispatchWithOrdering(event)
	}

	fmt.Println("P2P chat started")
	return p2pInstance.Host.ID().String()
}

func StartSubscription(callback PeerMessageCallback){
	// Listen for incoming messages
	p2pNode.EventManager.RegisterEventHandler("message", func(event EventMessage) {
		fmt.Println("Received event from %s: %s\n", event.SenderNick, event.Data)
		callback.OnMessage(event.Data)
		if err := storage.AddEventIfNotDuplicate(event); err != nil {
			fmt.Printf("Error adding event to storage: %v\n", err)
		}
	})

	// Periodic synchronization
	go storage.PeriodicSync(p2pNode.EventManager, p2pNode.Host.Peerstore().Peers(), p2pNode.Host, 30*time.Second)
}

func SubscribeToPeers(callback PeerCallback){
	go func() {
		for {
			time.Sleep(5 * time.Second)
			peers := p2pNode.Host.Peerstore().Peers()
			callback.OnMessage(peers.String())
		}
	}()
}

func PublishMessage(message string) error {
    return globalRoom.Publish(EventTypeMessage, message)
}