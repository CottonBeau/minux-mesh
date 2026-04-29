package core

import (
	"context"
	"fmt"
	"os"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	host host.Host
	ctx  context.Context
	cancel context.CancelFunc
}

func NewNode() (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	port := "4001"
	for i, arg := range os.Args[1:] {
		if arg == "--port" && i+1 < len(os.Args[1:]) {
			port = os.Args[i+2]
		}
	}

	priv, err := loadOrCreateIdentity("device.key")
	if err != nil {
		cancel()
		return nil, fmt.Errorf("identity error: %w", err)
	}

	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
			     libp2p.Transport(tcp.NewTCPTransport),
			     libp2p.Identity(priv),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("host error: %w", err)
	}

	h.SetStreamHandler("/minux/1.0.0", func(s network.Stream) {
		fmt.Println("New node connected:", s.Conn().RemotePeer())
		s.Write([]byte(h.ID().String() + "\n"))
		time.Sleep(5 * time.Second)
		s.Close()
	})

	fmt.Println("Node started")
	fmt.Println("Device ID:", h.ID())
	fmt.Printf("Listening on: %s/p2p/%s\n", h.Addrs()[0], h.ID())

	return &Node{host: h, ctx: ctx, cancel: cancel}, nil
}

func (n *Node) Run() {
	var peerAddr string
	for i, arg := range os.Args[1:] {
		if arg[:1] == "/" {
			peerAddr = arg
		} else if arg == "--port" {
			_ = i
		}
	}

	if peerAddr != "" {
		if err := n.ConnectToPeer(peerAddr); err != nil {
			fmt.Println("Error:", err)
		}
	}

	<-n.ctx.Done()
}

func (n *Node) Close() {
	n.cancel()
	n.host.Close()
}

func (n *Node) ConnectToPeer(peerAddr string) error {
	maddr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("invalid peer info: %w", err)
	}

	if err := n.host.Connect(n.ctx, *peerInfo); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	s, err := n.host.NewStream(n.ctx, peerInfo.ID, "/minux/1.0.0")
	if err != nil {
		return fmt.Errorf("stream failed: %w", err)
	}
	defer s.Close()

	fmt.Println("Connected to:", peerInfo.ID)
	return nil
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
