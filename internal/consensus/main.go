package consensus

import (
	"fmt"
	"log"
	"strconv"

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

	repo := transaction.New(folder+"/"+strconv.Itoa(port)+".result", n)
	usecase := consensus.New(repo, host, k, alpha, beta)

	host.SetHandler(usecase.Answer)
	host.SetDone(repo.Done)

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
		log.Println("[INFO] Decided transaction", i, ", go to the next transaction")
	}

	if err := a.repo.Commit(); err != nil {
		panic(err)
	}
	log.Println("[INFO] Commited")
}
