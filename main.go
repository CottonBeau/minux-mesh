package main

import (
	"fmt"
	"context"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func main() {
	ctx := context.Background()

	port := "4001"
	var peerAddr string

	for i, arg := range os.Args[1:] {
		if arg == "--port" && i+1 < len(os.Args[1:]) {
			port = os.Args[i+2]
		} else if arg[:1] == "/" {
			peerAddr = arg
		}
	}
	priv, err := loadOrCreateIdentity("device.key")
	if err != nil {
		panic(err)
	}
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/" + port),
				libp2p.Transport(tcp.NewTCPTransport),
				libp2p.Identity(priv),
	)
	if err != nil {
		panic(err)
	}
	defer host.Close()

	host.SetStreamHandler("/minux/1.0.0", func(s network.Stream) {
		fmt.Println("New node connected:", s.Conn().RemotePeer())
		s.Close()
	})

	fmt.Println("Node started")
	fmt.Println("Device ID:", host.ID())
	fmt.Printf("Listening on: %s/p2p/%s\n", host.Addrs()[0], host.ID())

	if peerAddr != "" {
		if err := connectToPeer(host, peerAddr); err != nil {
			fmt.Println("Error:", err)
		}
	}

	<-ctx.Done()
}
func loadOrCreateIdentity(path string) (crypto.PrivKey, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key: %w", err)
		}

		keyBytes, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal key: %w", err)
		}

		if err := os.WriteFile(path, keyBytes, 0600); err != nil {
			return nil, fmt.Errorf("failed to write key file: %w", err)
		}

		fmt.Println("New device identity created")
		return priv, nil
	}

	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	priv, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal key: %w", err)
	}

	fmt.Println("Device identity loaded")
	return priv, nil
}
func connectToPeer(host host.Host, peerAddr string) error {
	maddr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("invalid peer info: %w", err)
	}

	if err := host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	s, err := host.NewStream(context.Background(), peerInfo.ID, "/minux/1.0.0")
	if err != nil {
		return fmt.Errorf("stream failed: %w", err)
	}
	defer s.Close()

	fmt.Println("Connected to:", peerInfo.ID)
	return nil
}
