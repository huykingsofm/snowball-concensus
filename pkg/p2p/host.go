package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/huykingsofm/snowball-concensus/internal/entity"
)

type Host struct {
	listener net.Listener
	minPort  int
	maxPort  int
	orders   []int
	idxOrder int
	handler  func(ix uint) (entity.Transaction, error)
}

func New(host string, port, minPort, maxPort int) (*Host, error) {
	rand.Seed(time.Now().UnixNano())
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	h := &Host{
		listener: listener,
		minPort:  minPort,
		maxPort:  maxPort,
		orders:   rand.Perm(maxPort - minPort),
		idxOrder: 0,
	}
	go h.Serve()

	return h, nil
}

func (h *Host) SetHandler(f func(ix uint) (entity.Transaction, error)) {
	h.handler = f
}

func (h *Host) Close() {
	time.Sleep(600)
	h.listener.Close()
}

func (h *Host) Serve() error {
	for {
		conn, err := h.listener.Accept()
		if err != nil {
			return err
		}

		go h.handleConnection(conn)
	}
}

func (h *Host) Ask(k, ix uint) ([]entity.Transaction, error) {
	nAttemps := 3
	preferences := []entity.Transaction{}
	for len(preferences) < int(k) {
		p, err := h.askOne(uint64(ix))
		if err != nil {
			if nAttemps <= 0 {
				return nil, err
			}

			log.Println("[WARNING] Can ask peer:", err)
			nAttemps--
		} else {
			preferences = append(preferences, p)
		}

		time.Sleep(10 * time.Millisecond)
	}

	return preferences, nil
}

func (h *Host) askOne(ix uint64) (entity.Transaction, error) {
	if h.idxOrder >= len(h.orders) {
		h.idxOrder = 0
		h.orders = rand.Perm(h.maxPort - h.minPort)
	}

	port := h.orders[h.idxOrder] + h.minPort
	h.idxOrder++

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return entity.Transaction{}, err
	}

	if err := binary.Write(conn, binary.BigEndian, &ix); err != nil {
		return entity.Transaction{}, err
	}

	var value int64
	if err := binary.Read(conn, binary.BigEndian, &value); err != nil {
		return entity.Transaction{}, err
	}

	log.Println("[DEBUG] Asking peer", port, "at", ix, "returns", value)
	return entity.Transaction{Value: int(value)}, nil
}

func (h *Host) handleConnection(conn net.Conn) error {
	defer conn.Close()
	if h.handler == nil {
		return errors.New("not setup handler")
	}

	var ix uint64
	if err := binary.Read(conn, binary.BigEndian, &ix); err != nil {
		return err
	}

	tx, err := h.handler(uint(ix))
	if err != nil {
		log.Println("[WARNING] Got a error when answer transaction ", ix)
		return err
	}

	if err := binary.Write(conn, binary.BigEndian, int64(tx.Value)); err != nil {
		return err
	}

	return nil
}
