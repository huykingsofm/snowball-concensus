package p2p

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
)

type ConnectionPool struct {
	ports        []int
	connections  []net.Conn
	permutations []int
	idx          int
	lock         sync.Mutex
}

func newConnectionPool(selfPort, minPort, maxPort int) *ConnectionPool {
	pool := &ConnectionPool{}

	for port := minPort; port < maxPort; port++ {
		if port == selfPort {
			continue
		}

		addr := fmt.Sprintf("localhost:%d", port)
		c, err := net.Dial("tcp", addr)
		if err != nil {
			log.Println("[WARNING] Cannot dial to", addr)
			continue
		}

		pool.ports = append(pool.ports, port)
		pool.connections = append(pool.connections, c)
	}

	log.Println("[INFO] Connected to", len(pool.connections), "peers")
	pool.permutations = rand.Perm(len(pool.connections))
	pool.idx = 0

	return pool
}

func (pool *ConnectionPool) Random() (net.Conn, int) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if pool.idx >= len(pool.connections) {
		pool.permutations = rand.Perm(len(pool.connections))
		pool.idx = 0
	}

	idx := pool.permutations[pool.idx]
	c := pool.connections[idx]
	p := pool.ports[idx]
	pool.idx++

	return c, p
}

func (pool *ConnectionPool) Get(i int) (net.Conn, int) {
	return pool.connections[i], pool.ports[i]
}

func (pool *ConnectionPool) Len() int {
	return len(pool.connections)
}
