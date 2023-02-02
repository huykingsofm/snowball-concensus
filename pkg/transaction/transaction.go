package transaction

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/huykingsofm/snowball-concensus/internal/entity"
)

type ConflictTransaction struct {
	entity.Transaction
	consecutiveSuccess uint
	confidence         uint
}

type ConcensusTransaction struct {
	name     string
	decided  []entity.Transaction
	conflict [][]ConflictTransaction
}

func New(name string, nElements uint) *ConcensusTransaction {
	rand.Seed(time.Now().UnixNano())

	c := &ConcensusTransaction{
		name:     name,
		decided:  nil,
		conflict: make([][]ConflictTransaction, nElements),
	}

	for i := range c.conflict {
		c.conflict[i] = make([]ConflictTransaction, 1)
		c.conflict[i][0] = ConflictTransaction{
			Transaction: entity.Transaction{
				Value: rand.Intn(i+4) + (i - 4),
			},
			consecutiveSuccess: 0,
			confidence:         1,
		}
	}

	log.Println("[INFO] Conflict set:", c.conflict)

	return c
}

func (c *ConcensusTransaction) Commit() error {
	if len(c.decided) < len(c.conflict) {
		return fmt.Errorf("it has not yet decided all transaction")
	}

	f, err := os.OpenFile(c.name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	for i := range c.decided {
		f.WriteString(strconv.Itoa(c.decided[i].Value) + " ")
	}

	return nil
}

// Get returns the transaction with the highest confidence in ix'th transactions.
func (c *ConcensusTransaction) Get(ix uint) (entity.Transaction, error) {
	if int(ix) >= len(c.conflict) {
		return entity.Transaction{}, fmt.Errorf("invalid index %d", ix)
	}

	if int(ix) < len(c.decided) {
		return c.decided[ix], nil
	}

	tx := entity.Transaction{}
	maxConfidence := uint(0)
	for i := range c.conflict[ix] {
		if c.conflict[ix][i].confidence > maxConfidence {
			maxConfidence = c.conflict[ix][i].confidence
			tx = c.conflict[ix][i].Transaction
		}
	}

	if maxConfidence == 0 {
		return tx, errors.New("can not find any transaction in conflict set")
	}

	return tx, nil
}

// GetHighestSuccess returns the transaction with the highest consecutiveSuccess
// in ix'th transactions.
func (c *ConcensusTransaction) GetHighestSuccess(ix uint) (tx entity.Transaction, consecutiveSuccess uint, err error) {
	if int(ix) >= len(c.conflict) {
		return entity.Transaction{}, 0, fmt.Errorf("invalid index %d", ix)
	}

	if int(ix) < len(c.decided) {
		return entity.Transaction{}, 0, fmt.Errorf("the transaction %d is decided", ix)
	}

	for i := range c.conflict[ix] {
		if c.conflict[ix][i].consecutiveSuccess > consecutiveSuccess {
			consecutiveSuccess = c.conflict[ix][i].consecutiveSuccess
			tx = c.conflict[ix][i].Transaction
		}
	}

	return tx, consecutiveSuccess, nil
}

// Update sets tx as the success transaction, increase its confidence and consecutiveSuccess.
// Then its sets consecutiveSuccess of other ix'th transactions to zero.
func (c *ConcensusTransaction) Update(ix uint, tx entity.Transaction) error {
	if int(ix) >= len(c.conflict) {
		return fmt.Errorf("invalid index %d", ix)
	}

	if int(ix) < len(c.decided) {
		return fmt.Errorf("the transaction %d is decided", ix)
	}

	hasTx := false
	for i := range c.conflict[ix] {
		if c.conflict[ix][i].Transaction.Value == tx.Value {
			c.conflict[ix][i].confidence += 1
			c.conflict[ix][i].consecutiveSuccess += 1
			hasTx = true
		} else {
			c.conflict[ix][i].consecutiveSuccess = 0
		}
	}

	if !hasTx {
		c.conflict[ix] = append(c.conflict[ix], ConflictTransaction{
			Transaction:        tx,
			confidence:         1,
			consecutiveSuccess: 1,
		})
	}

	return nil
}

// UpdateFailed sets consecutiveSuccess of all ix'th transaction to zero.
func (c *ConcensusTransaction) UpdateFailed(ix uint) error {
	if int(ix) >= len(c.conflict) {
		return fmt.Errorf("invalid index %d", ix)
	}

	if int(ix) < len(c.decided) {
		return fmt.Errorf("the transaction %d is decided", ix)
	}

	for i := range c.conflict[ix] {
		c.conflict[ix][i].consecutiveSuccess = 0
	}

	return nil
}

// Decide chooses tx as the accepted ix'th transaction.
func (c *ConcensusTransaction) Decide(ix uint, tx entity.Transaction) error {
	if int(ix) >= len(c.conflict) {
		return fmt.Errorf("invalid index %d", ix)
	}

	if int(ix) < len(c.decided) {
		return fmt.Errorf("the transaction %d is decided", ix)
	}

	if int(ix) > len(c.decided) {
		return errors.New("do not decide transaction if previous one is not decided")
	}

	c.decided = append(c.decided, tx)
	return nil
}

func (c *ConcensusTransaction) Done() bool {
	return len(c.decided) == len(c.conflict)
}
