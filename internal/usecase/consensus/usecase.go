package consensus

import (
	"log"

	"github.com/huykingsofm/snowball-concensus/internal/entity"
)

type usecase struct {
	k     uint
	alpha uint
	beta  uint
	repo  ConsensusTransactionRepo
	peers PeersService
}

func New(repo ConsensusTransactionRepo, peers PeersService, k, alpha, beta uint) Usecase {
	u := usecase{
		repo: repo, peers: peers, k: k, alpha: alpha, beta: beta,
	}
	return u
}

func (u usecase) Decide(ix uint) error {
	for {
		preferences, err := u.peers.Ask(u.k, ix)
		if err != nil {
			return err
		}
		pCount := map[entity.Transaction]uint{}
		for _, tx := range preferences {
			if _, ok := pCount[tx]; !ok {
				pCount[tx] = 0
			}
			pCount[tx] += 1
		}

		var p entity.Transaction
		var m uint
		for k, v := range pCount {
			if v > m {
				m = v
				p = k
			}
		}

		if m >= u.alpha {
			log.Println("[INFO] Choose tx with value: ", p.Value)
			if err := u.repo.Update(ix, p); err != nil {
				return err
			}
		} else {
			log.Println("[INFO] No tx success, reset all consecutive success")
			if err := u.repo.UpdateFailed(ix); err != nil {
				return err
			}
		}

		tx, n, err := u.repo.GetHighestSuccess(ix)
		log.Println("[INFO] Highest consecutive success: value", tx.Value, "with", n, "times")
		if err != nil {
			return err
		}

		if n >= u.beta {
			u.repo.Decide(ix, tx)
			break
		}
	}

	return nil
}

func (u usecase) Answer(ix uint) (entity.Transaction, error) {
	return u.repo.Get(ix)
}
