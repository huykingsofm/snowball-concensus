package consensus

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"unsafe"

	"github.com/huykingsofm/snowball-concensus/internal/usecase/consensus"
	"github.com/huykingsofm/snowball-concensus/pkg/p2p"
	"github.com/huykingsofm/snowball-concensus/pkg/transaction"
	"github.com/xybor-x/xycond"
)

type App struct {
	host     *p2p.Host
	repo     *transaction.ConcensusTransaction
	nElement uint
	usecase  consensus.Usecase
}

func New(folder string, port, minPort, maxPort int, n, k, alpha, beta uint) App {
	if alpha > k {
		panic("alpha must not larger than k samples")
	}

	if port > maxPort || port < minPort {
		panic(fmt.Sprintf("invalid port %d not in %d-%d", port, minPort, maxPort))
	}

	host, err := p2p.New("localhost", port, minPort, maxPort)
	xycond.AssertNil(err)

	repo := transaction.New(folder+"/"+generateRandomString(10), n)
	usecase := consensus.New(repo, host, k, alpha, beta)

	host.SetHandler(usecase.Answer)

	return App{
		host:     host,
		repo:     repo,
		nElement: n,
		usecase:  usecase,
	}
}

func (a App) Run() {
	defer a.host.Close()

	i := uint(0)
	for i < a.nElement {
		if err := a.usecase.Decide(i); err != nil {
			log.Println("[WARNING] " + err.Error())
		} else {
			i++
		}
		log.Println("[INFO] Sleep sometime")
		time.Sleep(time.Second)
	}

	if err := a.repo.Commit(); err != nil {
		panic(err)
	}
	log.Println("[INFO] Commited")
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var randomSrc = rand.NewSource(time.Now().UnixNano())

func generateRandomString(n int) string {
	b := make([]byte, n)

	for i, cache, remain := n-1, randomSrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randomSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
