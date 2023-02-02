package consensus

import "github.com/huykingsofm/snowball-concensus/internal/entity"

type ConsensusTransactionRepo interface {
	// Get returns the transaction with the highest confidence in ix'th transactions.
	Get(ix uint) (entity.Transaction, error)

	// GetHighestSuccess returns the transaction with the highest consecutiveSuccess
	// in ix'th transactions.
	GetHighestSuccess(ix uint) (tx entity.Transaction, consecutiveSuccess uint, err error)

	// Update sets tx as the success transaction, increase its confidence and consecutiveSuccess.
	// Then its sets consecutiveSuccess of other ix'th transactions to zero.
	Update(ix uint, tx entity.Transaction) error

	// UpdateFailed sets consecutiveSuccess of all ix'th transaction to zero.
	UpdateFailed(ix uint) error

	// Decide chooses tx as the accepted ix'th transaction.
	Decide(ix uint, tx entity.Transaction) error
}

type PeersService interface {
	// Ask asks the k peer about their preference transaction.
	Ask(k, ix uint) ([]entity.Transaction, error)
}

type Usecase interface {
	// Decide will choose the consensus transaction at ix index.
	Decide(ix uint) error

	// Answer returns the preference transaction.
	Answer(ix uint) (entity.Transaction, error)
}
