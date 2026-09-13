package ktfunc

// Fee-reserve accounting.
//
// The contract books OC fees into tlOcFees on every vote and reward, but it
// never checks that its ETH balance still covers them: rwd sends whatever
// _amt the calling node passes in (less one more fee), and withdrawOCFee
// zeroes a claim it cannot pay instead of reverting. So the one number that
// says whether operators can actually be paid is balance − tlOcFees, and every
// reader of it lives here so -ktProps, the run loop, the reward audit and the
// withdraw guard all agree on the arithmetic and the wording.

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

const (
	FeeReserveOKLabel          = "OK"
	FeeReserveUnderfundedLabel = "UNDERFUNDED"
)

// FeeReserve is the contract's ETH balance set against the fees it owes.
type FeeReserve struct {
	Balance  *big.Int // contract ETH balance
	TlOcFees *big.Int // fees the contract owes its OCs (tlOcFees)
	Reserve  *big.Int // Balance − TlOcFees: the next pot when positive, a deficit when negative
}

// Underfunded reports whether the contract cannot cover the fees it owes.
func (r *FeeReserve) Underfunded() bool { return r.Reserve.Sign() < 0 }

// Deficit is how much the balance falls short of the fees owed (0 when whole).
func (r *FeeReserve) Deficit() *big.Int {
	if !r.Underfunded() {
		return big.NewInt(0)
	}
	return new(big.Int).Neg(r.Reserve)
}

// Pot is what a correctly computed reward would pass to rwd right now:
// balance − tlOcFees, floored at zero.
func (r *FeeReserve) Pot() *big.Int {
	if r.Reserve.Sign() < 0 {
		return big.NewInt(0)
	}
	return new(big.Int).Set(r.Reserve)
}

func (r *FeeReserve) String() string {
	if r.Underfunded() {
		return fmt.Sprintf("balance %s ETH - OC fees owed %s ETH = %s by %s ETH (winners get nothing until %s ETH of income refills the reserve)",
			fmtEth(r.Balance), fmtEth(r.TlOcFees), FeeReserveUnderfundedLabel, fmtEth(r.Deficit()), fmtEth(r.Deficit()))
	}
	return fmt.Sprintf("balance %s ETH - OC fees owed %s ETH = %s ETH next pot (%s)",
		fmtEth(r.Balance), fmtEth(r.TlOcFees), fmtEth(r.Reserve), FeeReserveOKLabel)
}

// ReadFeeReserve reads balance and tlOcFees at the given block (nil = latest).
// Historic blocks need an RPC that still holds that state.
func ReadFeeReserve(cProps *ConnectionProps, block *big.Int) (*FeeReserve, error) {
	ctx := context.Background()
	balance, err := cProps.Client.BalanceAt(ctx, cProps.KtAddr, block)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract balance at %s: %w", blockLabel(block), err)
	}
	owed, err := cProps.Kt.TlOcFees(&bind.CallOpts{Context: ctx, From: cProps.MyPubKey, BlockNumber: block})
	if err != nil {
		return nil, fmt.Errorf("failed to read tlOcFees at %s: %w", blockLabel(block), err)
	}
	return &FeeReserve{Balance: balance, TlOcFees: owed, Reserve: new(big.Int).Sub(balance, owed)}, nil
}

// FeeClaim is what withdrawOCFee would try to pay one OC right now. The
// contract pays it all-or-nothing and forfeits the claim when it cannot.
type FeeClaim struct {
	Past         *big.Int // pastOcFees[oc]: fees already migrated from earlier epochs
	Unmigrated   *big.Int // ocFees[oc][lastStartBlock], folded in by migrateFees on the next call
	CurrentEpoch *big.Int // ocFees[oc][startBlock], paid only once the epoch is complete
	Total        *big.Int
}

// EstimateFeeClaim reproduces the contract's withdrawOCFee arithmetic for oc:
// migrateFees first moves the fee parked under the OC's lastStartBlock into
// pastOcFees (when that epoch is older than the current one), then the
// current epoch's fee is added if the epoch is complete, and the sum is paid.
func EstimateFeeClaim(cProps *ConnectionProps, oc common.Address) (*FeeClaim, error) {
	ctx := context.Background()
	opts := &bind.CallOpts{Context: ctx, From: cProps.MyPubKey}

	past, err := cProps.Kt.PastOcFees(opts, oc)
	if err != nil {
		return nil, fmt.Errorf("failed to query pastOcFees for %s: %w", oc.Hex(), err)
	}
	startBlock, err := cProps.Kt.StartBlock(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to read start block: %w", err)
	}
	interval, err := cProps.Kt.EpochInterval(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to read epoch interval: %w", err)
	}
	head, err := cProps.Client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read current block: %w", err)
	}
	lastStart, err := cProps.Kt.LastStartBlock(opts, oc)
	if err != nil {
		return nil, fmt.Errorf("failed to read lastStartBlock for %s: %w", oc.Hex(), err)
	}

	claim := &FeeClaim{Past: past, Unmigrated: big.NewInt(0), CurrentEpoch: big.NewInt(0)}
	if lastStart.Sign() != 0 && lastStart.Cmp(startBlock) < 0 {
		unmigrated, err := cProps.Kt.OcFees(opts, oc, lastStart)
		if err != nil {
			return nil, fmt.Errorf("failed to read ocFees for %s at epoch %s: %w", oc.Hex(), lastStart, err)
		}
		claim.Unmigrated = unmigrated
	}
	// The contract tests block.number > startBlock + interval when the
	// withdraw tx executes, i.e. at head+1, so the epoch counts as complete
	// once head has reached its end block.
	if head >= startBlock.Uint64()+uint64(interval) {
		current, err := cProps.Kt.OcFees(opts, oc, startBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to read ocFees for %s at epoch %s: %w", oc.Hex(), startBlock, err)
		}
		claim.CurrentEpoch = current
	}
	claim.Total = new(big.Int).Add(new(big.Int).Add(claim.Past, claim.Unmigrated), claim.CurrentEpoch)
	return claim, nil
}

// PrintFeeReserve logs the reserve line and this node's withdrawable claim,
// with a warning when withdrawing now would forfeit it.
func PrintFeeReserve(cProps *ConnectionProps) error {
	reserve, err := ReadFeeReserve(cProps, nil)
	if err != nil {
		return err
	}
	if reserve.Underfunded() {
		log.Warnf("Fee reserve: %s", reserve)
		log.Warnf("  A reward overpaid into the reserve. Run -%s %s to find which OC sent it.", AuditRewardsFlagName, AuditRangeAll)
	} else {
		log.Printf("Fee reserve: %s", reserve)
	}

	claim, err := EstimateFeeClaim(cProps, cProps.MyPubKey)
	if err != nil {
		return err
	}
	log.Printf("This node's fee claim: %s ETH (migrated %s, awaiting migration %s, current epoch %s)",
		fmtEth(claim.Total), fmtEth(claim.Past), fmtEth(claim.Unmigrated), fmtEth(claim.CurrentEpoch))
	if claim.Total.Sign() > 0 && reserve.Balance.Cmp(claim.Total) < 0 {
		log.Warnf("  Do NOT run -withdrawFees yet: the contract holds %s ETH, less than this claim, and withdrawOCFee would zero the claim without paying it.",
			fmtEth(reserve.Balance))
	}
	return nil
}

// fmtEth renders wei as a fixed 6-decimal ETH string for log lines.
func fmtEth(wei *big.Int) string {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18)).Text('f', 6)
}

func blockLabel(block *big.Int) string {
	if block == nil {
		return "latest"
	}
	return "block " + block.String()
}
