package ktfunc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	log "github.com/sirupsen/logrus"
)

// Ktv2Filterer is the interface for event filtering.
type Ktv2Filterer interface {
	FilterStaked(opts *bind.FilterOpts) (*Ktv2StakedIterator, error)
	FilterWithdrew(opts *bind.FilterOpts) (*Ktv2WithdrewIterator, error)
}

// Ktv2Staked represents a Staked event.
type Ktv2Staked struct {
	Arg0 common.Address
	Arg1 *big.Int
	Raw  types.Log
}

// Ktv2Withdrew represents a Withdrew event.
type Ktv2Withdrew struct {
	Arg0 common.Address
	Arg1 *big.Int
	Raw  types.Log
}

type Ktv2StakedIterator struct {
	Event *Ktv2Staked
	err   error
	logs  chan types.Log
	sub   event.Subscription
}

// Next advances the iterator to the next event.
func (it *Ktv2StakedIterator) Next() bool {
	select {
	case log, ok := <-it.logs:
		if !ok {
			return false
		}
		it.Event = &Ktv2Staked{Raw: log}
		// Parse log data (address and uint256)
		if len(log.Data) >= 64 {
			it.Event.Arg0 = common.BytesToAddress(log.Data[0:32])
			it.Event.Arg1 = new(big.Int).SetBytes(log.Data[32:64])
		}
		return true
	case err, ok := <-it.sub.Err():
		if !ok {
			return false
		}
		it.err = err
		return false
	}
}

// Error returns the error from the iterator.
func (it *Ktv2StakedIterator) Error() error {
	return it.err
}

// Close closes the iterator.
func (it *Ktv2StakedIterator) Close() error {
	it.sub.Unsubscribe()
	close(it.logs)
	return nil
}

type Ktv2WithdrewIterator struct {
	Event *Ktv2Withdrew
	err   error
	logs  chan types.Log
	sub   event.Subscription
}

// Next advances the iterator to the next event.
func (it *Ktv2WithdrewIterator) Next() bool {
	select {
	case log, ok := <-it.logs:
		if !ok {
			return false
		}
		it.Event = &Ktv2Withdrew{Raw: log}
		// Parse log data (address and uint256)
		if len(log.Data) >= 64 {
			it.Event.Arg0 = common.BytesToAddress(log.Data[0:32])
			it.Event.Arg1 = new(big.Int).SetBytes(log.Data[32:64])
		}
		return true
	case err, ok := <-it.sub.Err():
		if !ok {
			return false
		}
		it.err = err
		return false
	}
}

// Error returns the error from the iterator.
func (it *Ktv2WithdrewIterator) Error() error {
	return it.err
}

// Close closes the iterator.
func (it *Ktv2WithdrewIterator) Close() error {
	it.sub.Unsubscribe()
	close(it.logs)
	return nil
}

// PrintFilteredEvents fetches and prints Staked and Withdrew events with debug logging.
func PrintFilteredStakeEvents(cProps *ConnectionProps, startBlock, endBlock uint64) error {
	log.Infof("Filtering events from block %d to %d for contract %s", startBlock, endBlock, cProps.KtAddr.Hex())

	opts := &bind.FilterOpts{
		Start:   startBlock,
		End:     &endBlock,
		Context: context.Background(),
	}

	// Filter Staked events
	stakeIter, err := cProps.Kt.FilterStaked(opts)
	if err != nil {
		log.Errorf("Failed to filter Staked events: %v", err)
		return fmt.Errorf("failed to filter Staked events: %w", err)
	}
	log.Info("Staked Events:")
	for stakeIter.Next() {
		e := stakeIter.Event()
		if e == nil {
			log.Debug("Nil Staked event")
			continue
		}
		log.Infof("Block: %d, Address: %s, Amount: %s", e.Raw.BlockNumber, e.Arg0.Hex(), e.Arg1.String())
	}
	if err := stakeIter.Error(); err != nil {
		log.Errorf("Stake iterator error: %v", err)
	}
	stakeIter.Close()

	// Filter Withdrew events
	withdrawIter, err := cProps.Kt.FilterWithdrew(opts)
	if err != nil {
		log.Errorf("Failed to filter Withdrew events: %v", err)
		return fmt.Errorf("failed to filter Withdrew events: %w", err)
	}
	log.Info("Withdrew Events:")
	for withdrawIter.Next() {
		e := withdrawIter.Event()
		if e == nil {
			log.Debug("Nil Withdrew event")
			continue
		}
		log.Infof("Block: %d, Address: %s, Amount: %s", e.Raw.BlockNumber, e.Arg0.Hex(), e.Arg1.String())
	}
	if err := withdrawIter.Error(); err != nil {
		log.Errorf("Withdraw iterator error: %v", err)
	}
	withdrawIter.Close()

	return nil
}
