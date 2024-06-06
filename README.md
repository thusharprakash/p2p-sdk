Sure, here's a comprehensive README for the `p2p` package:

---

# p2p Package

The `p2p` package provides an easy-to-use library for building peer-to-peer (P2P) applications using libp2p. This package encapsulates the core functionalities needed for creating a P2P network, setting up pubsub messaging, and handling mDNS peer discovery.

## Features

- Peer-to-peer communication using libp2p
- Pubsub messaging with GossipSub
- Local peer discovery using mDNS
- Easy-to-use API for joining rooms and publishing messages

## Installation

To use the `p2p` package in your project, you need to import it and ensure you have the necessary dependencies. Add the following to your `go.mod` file:

```go
module p2p-sdk

go 1.16

require (
    github.com/libp2p/go-libp2p v0.19.1
    github.com/libp2p/go-libp2p-core v0.9.0
    github.com/libp2p/go-libp2p-pubsub v0.4.1
    github.com/libp2p/go-libp2p-discovery v0.5.1
)
```

Then run:
```sh
go mod tidy
```

## Usage

### Importing the Package

First, import the `p2p` package in your application:

```go
import "p2p-sdk/p2p"
```

### Initializing the P2P Library

Initialize the P2P library and join a room:

```go
ctx := context.Background()

// Initialize the P2P library
p2pInstance, err := p2p.NewP2P(ctx)
if err != nil {
    panic(err)
}

// Join an event room
nick := "your-nickname"
roomName := "example-room"

room, err := p2pInstance.JoinRoom(ctx, roomName, nick)
if err != nil {
    panic(err)
}
```

### Publishing Messages

Publish a message to the room:

```go
if err := room.Publish("Hello, world!"); err != nil {
    fmt.Printf("Error publishing message: %s\n", err)
}
```

### Receiving Messages

Listen for incoming messages in a separate goroutine:

```go
go func() {
    for msg := range room.Messages {
        fmt.Printf("[%s] %s: %s\n", roomName, msg.SenderNick, msg.Message)
    }
}()
```

### Complete Example

Here's a complete example of using the `p2p` package in a chat application:

```go
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

    go func() {
        for msg := range room.Messages {
            fmt.Printf("[%s] %s: %s\n", roomName, msg.SenderNick, msg.Message)
        }
    }()

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

    select {}
}
```

## Directory Structure

The package is organized into the following files:

- `p2p.go`: Main initialization and room management.
- `pubsub.go`: Pubsub messaging functions.
- `mdns.go`: mDNS peer discovery functions.
- `utils.go`: Utility functions.

## Functions

### NewP2P

`NewP2P(ctx context.Context) (*P2P, error)`

Initializes a new P2P instance with a libp2p host and pubsub system.

### JoinRoom

`(p2p *P2P) JoinRoom(ctx context.Context, roomName, nick string) (*EventRoom, error)`

Joins a pubsub room with the given name and nickname.

### Publish

`(room *EventRoom) Publish(message string) error`

Publishes a message to the pubsub topic.

### ListPeers

`(room *EventRoom) ListPeers() []peer.ID`

Lists peers currently in the pubsub topic.

### SetupDiscovery

`SetupDiscovery(h host.Host) error`

Sets up mDNS peer discovery for the given host.

### Utility Functions

- `PrintErr(m string, args ...interface{})`
- `DefaultNick(p peer.ID) string`
- `ShortID(p peer.ID) string`

## License

This package is open-source and available under the MIT License.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

---

This README provides a comprehensive overview of the `p2p` package, its installation, usage, and directory structure. Feel free to adjust it based on your specific needs and project details.