package p2p

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/huykingsofm/snowball-concensus/internal/entity"
)

type Host struct {
	listener    net.Listener
	port        int
	minPort     int
	maxPort     int
	pool        *ConnectionPool
	handler     func(ix uint) (entity.Transaction, error)
	doneHandler func() bool
}

func New(host string, port int, minPort, maxPort int) *Host {
	return &Host{
		port:    port,
		minPort: minPort,
		maxPort: maxPort,
	}
}

func (h *Host) SetHandler(f func(ix uint) (entity.Transaction, error)) {
	h.handler = f
}

func (h *Host) SetDone(f func() bool) {
	h.doneHandler = f
}

func (h *Host) Close() {
	nConn := h.pool.Len()
	for i := 0; i < nConn; i++ {
		conn, port := h.pool.Get(i)
		for {
			log.Println("[INFO] Wait for process", port, "commits...")
			if h.askDone(conn) {
				conn.Close()
				break
			}
			time.Sleep(time.Second)
		}
	}

	log.Println("[INFO] Close the connections")
	h.listener.Close()
}

func (h *Host) ListenAndAccept() error {
	rand.Seed(time.Now().UnixNano())
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", h.port))
	if err != nil {
		return err
	}

	h.listener = listener
	go func() {
		for {
			conn, err := h.listener.Accept()
			if err != nil {
				log.Println("[WARNING] Cannot accept connection:", err)
				return
			}
			go h.handleConnection(conn)
		}
	}()

	time.Sleep(2 * time.Second)
	h.pool = newConnectionPool(h.port, h.minPort, h.maxPort)

	return nil
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

			log.Println("[WARNING] Can not ask peer:", err)
			nAttemps--
		} else {
			preferences = append(preferences, p)
		}
	}

	return preferences, nil
}

func (h *Host) askOne(ix uint64) (entity.Transaction, error) {
	conn, port := h.pool.Random()

	if err := binary.Write(conn, binary.BigEndian, &ix); err != nil {
		return entity.Transaction{}, err
	}

	var value int64
	if err := binary.Read(conn, binary.BigEndian, &value); err != nil {
		return entity.Transaction{}, err
	}

	log.Println("[DEBUG] Asking peer", port, "at index", ix, "returns", value)
	return entity.Transaction{Value: int(value)}, nil
}

func (h *Host) askDone(conn net.Conn) bool {
	value := int64(-1)
	if err := binary.Write(conn, binary.BigEndian, value); err != nil {
		return true
	}

	if err := binary.Read(conn, binary.BigEndian, &value); err != nil {
		return true
	}

	return value == 1
}

func (h *Host) handleConnection(conn net.Conn) {
	for {
		var ix int64
		if err := binary.Read(conn, binary.BigEndian, &ix); err != nil {
			log.Println("[WARNING] Cannot read from conn:", err)
			conn.Close()
			break
		}

		if ix < 0 {
			result := int64(0)
			if h.doneHandler() {
				result = 1
			}
			if err := binary.Write(conn, binary.BigEndian, result); err != nil {
				log.Println("[WARNING] Cannot write done to conn:", err)
			}
			continue
		}

		if h.handler == nil {
			log.Println("[WARNING] Not setup handler, not handle connection")
		}

		tx, err := h.handler(uint(ix))
		if err != nil {
			log.Println("[WARNING] Got a error when answer transaction ", ix)
			continue
		}

		if err := binary.Write(conn, binary.BigEndian, int64(tx.Value)); err != nil {
			log.Println("[WARNING] Cannot write transaction to conn:", err)
			continue
		}

	}
}
