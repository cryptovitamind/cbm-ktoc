package ktfunc

// Reward audit.
//
// rwd(_to, _amt) trusts the amount the calling node passes in, so a node
// running an old build (before fees were subtracted from the pot) sends the
// contract's whole balance to the winner and drains the fee reserve that the
// other operators are owed. The Rwd event only records what was paid, not
// what should have been, so the audit rebuilds "should have been" from state
// at the block before each reward: balance − tlOcFees. Anything above that is
// an overpayment, attributed to the OC that signed the transaction.

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	AuditRewardsFlagName = "auditRewards"
	// AuditRangeAll audits every reward from contract creation to the chain head.
	AuditRangeAll = "all"

	RewardVerdictOK        = "OK"
	RewardVerdictOverpaid  = "OVERPAID"
	RewardVerdictUnderpaid = "UNDERPAID"
	RewardVerdictUnknown   = "UNKNOWN"

	// feeReserveSince is the earliest versioned release guaranteed to subtract
	// tlOcFees from the pot before rewarding. The fix itself (March 2026)
	// predates the versioned banner, so any build that still passes the whole
	// balance to rwd is older than this.
	feeReserveSince = "v0.4.5-beta"
)

// RewardAudit is one Rwd event checked against the state it was computed from.
type RewardAudit struct {
	Block  uint64
	TxHash common.Hash
	Caller common.Address // OC that sent the rwd transaction
	Winner common.Address
	AmtArg *big.Int // _amt the node passed to rwd
	Paid   *big.Int // amount the Rwd event says was sent (AmtArg minus the rwd fee)

	Before   *FeeReserve // state at Block−1; nil when the RPC has no historic state
	After    *FeeReserve // state at Block; nil when unavailable
	Expected *big.Int    // Before.Pot(): what a current build would have passed
	Overpaid *big.Int    // AmtArg − Expected when positive, else 0

	// FullBalanceSend is the fingerprint of a build older than feeReserveSince:
	// _amt equal to the whole contract balance while fees were owed.
	FullBalanceSend bool
	Verdict         string
	Note            string // why the verdict is UNKNOWN, or other detail
}

// AuditRewards checks every Rwd event in [from, to].
func AuditRewards(cProps *ConnectionProps, from, to uint64) ([]RewardAudit, error) {
	return nil, nil
}

// ResolveAuditRange turns the -auditRewards argument into a block range:
// AuditRangeAll spans contract creation to head; otherwise <start>:<end>.
func ResolveAuditRange(cProps *ConnectionProps, spec string) (from, to uint64, err error) {
	return 0, 0, nil
}

// PrintRewardAudit logs one line per reward plus a summary and, for every OC
// seen sending a reward, its current fee claim.
func PrintRewardAudit(cProps *ConnectionProps, audits []RewardAudit) {}

// checkEpochAdvance runs from the vote loop each cycle. When startBlock has
// moved since the previous cycle, some node rewarded the epoch: audit the
// reward(s) that landed since the old epoch end and log the fee reserve.
func checkEpochAdvance(cProps *ConnectionProps, startBlock *big.Int, interval uint16, head uint64) {
}
