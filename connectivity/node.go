package connectivity

import (
	"context"
	"crypto/rand"
	"log/slog"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
)

func MakeNode(ctx context.Context) (host.Host, *pubsub.PubSub, error) {
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		slog.Error("Failed creating random key!", "error", err)
		return nil, nil, err
	}

	var dhtInst *dht.IpfsDHT
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.Muxer("/yamux/1.0.0", nil),
		libp2p.NATPortMap(),
		libp2p.EnableNATService(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var err error
			dhtInst, err = dht.New(ctx, h, dht.Mode(dht.ModeAuto))
			return dhtInst, err
		}),
	)
	if err != nil {
		slog.Error("Failed setting up a new libp2p client", "error", err)
		return nil, nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		slog.Error("Failed setting up gossip subscription", "error", err)
		return nil, nil, err
	}

	return h, ps, nil
}
