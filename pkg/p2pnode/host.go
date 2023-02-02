package p2pnode

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/huykingsofm/snowball-concensus/internal/entity"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
)

const protoID = protocol.ID("/consensus/1.0.0")

type Host struct {
	host    host.Host
	handler func(ix uint) (entity.Transaction, error)
}

func NewHost() (*Host, error) {
	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
	)

	muxers := libp2p.ChainOptions(
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
	)

	security := libp2p.Security(tls.ID, tls.New)

	listenAddrs := libp2p.ListenAddrStrings(
		"/ip4/0.0.0.0/tcp/0",
		"/ip4/0.0.0.0/tcp/0/ws",
	)

	var dht *kaddht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		dht, err = kaddht.New(context.Background(), h)
		return dht, err
	}
	routing := libp2p.Routing(newDHT)

	host, err := libp2p.New(
		transports,
		listenAddrs,
		muxers,
		security,
		routing,
	)
	if err != nil {
		return &Host{}, fmt.Errorf("can not start node %v", err)
	}

	h := &Host{host: host}
	h.host.SetStreamHandler(protoID, h.handle)

	mdns := mdns.NewMdnsService(host, "", h)
	if err := mdns.Start(); err != nil {
		return &Host{}, err
	}

	err = dht.Bootstrap(context.Background())
	if err != nil {
		return &Host{}, err
	}

	return h, nil
}

func (h *Host) SetHandler(f func(ix uint) (entity.Transaction, error)) {
	h.handler = f
}

func (m *Host) HandlePeerFound(pi peer.AddrInfo) {
	if m.host.Network().Connectedness(pi.ID) != network.Connected {
		fmt.Printf("Found %s!\n", pi.ID.ShortString())
		if err := m.host.Connect(context.Background(), pi); err != nil {
			log.Println("[WARNING] can not connect to " + pi.ID.ShortString())
		}
	}
}

func (h *Host) handle(s network.Stream) {
	var ix uint64
	binary.Read(s, binary.BigEndian, &ix)

	tx, err := h.handler(uint(ix))
	if err != nil {
		log.Println("[WARNING] " + err.Error())
	}

	if err := binary.Write(s, binary.BigEndian, int64(tx.Value)); err != nil {
		log.Println("[WARNING] " + err.Error())
	}
}

func (h *Host) Close() error {
	return h.host.Close()
}

func (h *Host) Ask(k, ix uint) ([]entity.Transaction, error) {
	preferences := map[peer.ID]entity.Transaction{}
	var peers []peer.ID
	var orders []int
	var index = 0
	for len(preferences) < int(k) {
		if index >= len(peers) {
			peers = h.host.Network().Peers()
			orders = rand.Perm(len(peers))
			index = 0
		}

		if len(peers) == 0 {
			return nil, fmt.Errorf("no peer found")
		}

		peer := peers[orders[index]]
		index++

		if _, ok := preferences[peer]; ok {
			continue
		}

		if _, err := h.host.Peerstore().SupportsProtocols(peer, string(protoID)); err != nil {
			log.Println("[WARNING] " + err.Error())
			continue
		}

		p, err := h.AskOne(peer, ix)
		if err != nil {
			log.Println("[WARNING] " + err.Error())
			continue
		}

		preferences[peer] = p
		time.Sleep(time.Second)
	}

	result := []entity.Transaction{}
	for _, v := range preferences {
		result = append(result, v)
	}

	return result, nil
}

func (h *Host) AskOne(peerID peer.ID, ix uint) (entity.Transaction, error) {
	s, err := h.host.NewStream(context.Background(), peerID, protoID)
	if err != nil {
		return entity.Transaction{}, err
	}

	if err := binary.Write(s, binary.BigEndian, uint64(ix)); err != nil {
		return entity.Transaction{}, err
	}

	var value int64
	if err := binary.Read(s, binary.BigEndian, &value); err != nil {
		return entity.Transaction{}, err
	}

	log.Println("[DEBUG] Asking", peerID.ShortString(), "at", ix, "returns", value)
	return entity.Transaction{Value: int(value)}, nil
}
