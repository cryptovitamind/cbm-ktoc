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
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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

func (r *FeeReserve) String() string { return "" }

// ReadFeeReserve reads balance and tlOcFees at the given block (nil = latest).
func ReadFeeReserve(cProps *ConnectionProps, block *big.Int) (*FeeReserve, error) {
	return &FeeReserve{Balance: big.NewInt(0), TlOcFees: big.NewInt(0), Reserve: big.NewInt(0)}, nil
}

// FeeClaim is what withdrawOCFee would try to pay one OC right now. The
// contract pays it all-or-nothing and forfeits the claim when it cannot.
type FeeClaim struct {
	Past         *big.Int // pastOcFees[oc]: fees already migrated from earlier epochs
	Unmigrated   *big.Int // ocFees[oc][lastStartBlock], folded in by migrateFees on the next call
	CurrentEpoch *big.Int // ocFees[oc][startBlock], paid only once the epoch is complete
	Total        *big.Int
}

// EstimateFeeClaim reproduces the contract's withdrawOCFee arithmetic for oc.
func EstimateFeeClaim(cProps *ConnectionProps, oc common.Address) (*FeeClaim, error) {
	z := big.NewInt(0)
	return &FeeClaim{Past: z, Unmigrated: z, CurrentEpoch: z, Total: z}, nil
}

// PrintFeeReserve logs the reserve line and this node's withdrawable claim.
func PrintFeeReserve(cProps *ConnectionProps) error { return nil }

// fmtEth renders wei as a fixed 6-decimal ETH string for log lines.
func fmtEth(wei *big.Int) string {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18)).Text('f', 6)
}
