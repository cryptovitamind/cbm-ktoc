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
	"bytes"
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"ktp2/src/abis/ktv2"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
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

	rwdMethodName = "rwd"
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

// AuditRewards checks every Rwd event in [from, to], in block order.
func AuditRewards(cProps *ConnectionProps, from, to uint64) ([]RewardAudit, error) {
	if to < from {
		return nil, fmt.Errorf("invalid audit range %d-%d", from, to)
	}
	parsed, err := ktv2.Ktv2MetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to load contract ABI: %w", err)
	}
	rwdMethod, ok := parsed.Methods[rwdMethodName]
	if !ok {
		return nil, fmt.Errorf("contract ABI has no %s method", rwdMethodName)
	}

	chunkSize := uint64(cProps.ChunkSize)
	if chunkSize == 0 {
		chunkSize = uint64(DefaultChunkSize)
	}

	audits := make([]RewardAudit, 0)
	for start := from; start <= to; {
		end := start + chunkSize - 1
		if end > to || end < start { // clamp, and guard the uint64 wrap
			end = to
		}
		iter, err := cProps.Kt.FilterRwd(&bind.FilterOpts{Start: start, End: &end, Context: context.Background()})
		if err != nil {
			// Same provider-limit fallback as the fee scans: shrink the window
			// and retry this chunk.
			if isLogLimitError(err) && chunkSize >= 2000 {
				chunkSize /= 2
				continue
			}
			return nil, fmt.Errorf("failed to filter Rwd events %d-%d: %w", start, end, err)
		}
		for iter.Next() {
			evt := iter.Event()
			if evt == nil {
				continue
			}
			audits = append(audits, auditReward(cProps, evt, rwdMethod))
		}
		if err := iter.Error(); err != nil {
			iter.Close()
			return nil, fmt.Errorf("error iterating Rwd events: %w", err)
		}
		iter.Close()
		if end == to {
			break
		}
		start = end + 1
	}
	sort.SliceStable(audits, func(i, j int) bool { return audits[i].Block < audits[j].Block })
	return audits, nil
}

func isLogLimitError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "query returned more than") || strings.Contains(msg, "log limit")
}

// auditReward rebuilds what one reward should have paid. Every step that can
// fail leaves the audit UNKNOWN with the reason, but keeps whatever was
// already recovered (caller, amount) so the operator still learns who sent it.
func auditReward(cProps *ConnectionProps, evt *ktv2.Ktv2Rwd, rwdMethod abi.Method) RewardAudit {
	ctx := context.Background()
	a := RewardAudit{
		Block:   evt.Raw.BlockNumber,
		TxHash:  evt.Raw.TxHash,
		Winner:  evt.Arg0,
		Paid:    evt.Arg1,
		Verdict: RewardVerdictUnknown,
	}

	tx, isPending, err := cProps.Client.TransactionByHash(ctx, evt.Raw.TxHash)
	if err != nil || isPending || tx == nil {
		a.Note = fmt.Sprintf("could not load reward transaction %s: %v", evt.Raw.TxHash.Hex(), err)
		return a
	}
	sender, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		a.Note = fmt.Sprintf("could not recover the sender of %s: %v", evt.Raw.TxHash.Hex(), err)
		return a
	}
	a.Caller = sender

	amt, err := decodeRwdAmount(rwdMethod, tx.Data())
	if err != nil {
		a.Note = fmt.Sprintf("could not decode rwd calldata: %v", err)
		return a
	}
	a.AmtArg = amt

	before, err := ReadFeeReserve(cProps, new(big.Int).SetUint64(a.Block-1))
	if err != nil {
		a.Note = fmt.Sprintf("no state at block %d (an archive-capable RPC is needed to judge this reward): %v", a.Block-1, err)
		return a
	}
	a.Before = before
	if after, err := ReadFeeReserve(cProps, new(big.Int).SetUint64(a.Block)); err == nil {
		a.After = after
	} else {
		a.Note = fmt.Sprintf("no state at block %d: %v", a.Block, err)
	}

	a.Expected = before.Pot()
	diff := new(big.Int).Sub(amt, a.Expected)
	a.Overpaid = big.NewInt(0)
	switch {
	case diff.Sign() > 0:
		a.Overpaid = diff
		a.Verdict = RewardVerdictOverpaid
	case diff.Sign() < 0:
		// A give landed between the node's read and the reward being mined.
		a.Verdict = RewardVerdictUnderpaid
	default:
		a.Verdict = RewardVerdictOK
	}
	a.FullBalanceSend = amt.Sign() > 0 && amt.Cmp(before.Balance) == 0 && before.TlOcFees.Sign() > 0
	return a
}

// decodeRwdAmount extracts _amt from rwd(address,uint256) calldata.
func decodeRwdAmount(rwdMethod abi.Method, data []byte) (*big.Int, error) {
	if len(data) < 4 || !bytes.Equal(data[:4], rwdMethod.ID) {
		return nil, fmt.Errorf("transaction is not a direct %s call", rwdMethodName)
	}
	args, err := rwdMethod.Inputs.Unpack(data[4:])
	if err != nil {
		return nil, err
	}
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	amt, ok := args[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("_amt is not a uint256")
	}
	return amt, nil
}

// ResolveAuditRange turns the -auditRewards argument into a block range:
// AuditRangeAll spans contract creation to head; otherwise <start>:<end>.
func ResolveAuditRange(cProps *ConnectionProps, spec string) (from, to uint64, err error) {
	if spec == AuditRangeAll {
		from, err = GetContractCreationBlock(cProps)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to find contract creation block: %w", err)
		}
		to, err = cProps.Client.BlockNumber(context.Background())
		if err != nil {
			return 0, 0, fmt.Errorf("failed to read current block: %w", err)
		}
		return from, to, nil
	}
	return ParseStartEndBlocks(spec)
}

// PrintRewardAudit logs one line per reward plus a summary and, for every OC
// seen sending a reward, whether it is still registered and what it is owed.
func PrintRewardAudit(cProps *ConnectionProps, audits []RewardAudit) {
	if len(audits) == 0 {
		log.Printf("No rewards found in the requested range.")
		return
	}
	overpaid, unknown := 0, 0
	callers := make([]common.Address, 0)
	seen := make(map[common.Address]bool)
	for _, a := range audits {
		logRewardAudit(a, cProps.MyPubKey)
		switch a.Verdict {
		case RewardVerdictOverpaid:
			overpaid++
		case RewardVerdictUnknown:
			unknown++
		}
		if a.Caller != (common.Address{}) && !seen[a.Caller] {
			seen[a.Caller] = true
			callers = append(callers, a.Caller)
		}
	}

	log.Printf("%d rewards audited: %d %s, %d %s", len(audits), overpaid, RewardVerdictOverpaid, unknown, RewardVerdictUnknown)
	if overpaid > 0 {
		log.Errorf("An %s reward passed more than balance - tlOcFees to rwd. A full-balance send means that OC runs a build older than %s and must upgrade.",
			RewardVerdictOverpaid, feeReserveSince)
	}
	if unknown > 0 {
		log.Warnf("%d rewards could not be judged; point ETH_ENDPOINT at an archive-capable RPC to audit old blocks.", unknown)
	}

	opts := &bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey}
	log.Printf("OCs that have sent rewards:")
	for _, oc := range callers {
		status := "registered"
		if active, err := cProps.Kt.OcRwdrs(opts, oc); err == nil && !active {
			status = "REMOVED"
		}
		past, err := cProps.Kt.PastOcFees(opts, oc)
		if err != nil {
			log.Printf("  %s  %s  (pastOcFees unavailable: %v)", oc.Hex(), status, err)
			continue
		}
		log.Printf("  %s  %s  pastOcFees %s ETH", oc.Hex(), status, fmtEth(past))
	}
	if reserve, err := ReadFeeReserve(cProps, nil); err == nil {
		log.Printf("Fee reserve now: %s", reserve)
	}
}

// logRewardAudit prints one audited reward at a level matching its verdict.
func logRewardAudit(a RewardAudit, me common.Address) {
	caller := a.Caller.Hex()
	if a.Caller == me {
		caller = "this node " + caller
	}
	switch a.Verdict {
	case RewardVerdictOverpaid:
		hint := ""
		if a.FullBalanceSend {
			hint = fmt.Sprintf(" It passed the whole contract balance: a build older than %s that does not reserve OC fees. That operator must upgrade.", feeReserveSince)
		}
		log.Errorf("Reward at block %d by OC %s: %s by %s ETH (passed %s ETH, should have been %s ETH; paid %s ETH to %s).%s",
			a.Block, caller, RewardVerdictOverpaid, fmtEth(a.Overpaid), fmtEth(a.AmtArg), fmtEth(a.Expected), fmtEth(a.Paid), a.Winner.Hex(), hint)
		if a.After != nil && a.After.Underfunded() {
			log.Errorf("  Fee reserve after it: %s", a.After)
		}
	case RewardVerdictUnderpaid:
		log.Printf("Reward at block %d by OC %s: %s (passed %s ETH, %s ETH was available; income landed between the read and the reward, harmless).",
			a.Block, caller, RewardVerdictUnderpaid, fmtEth(a.AmtArg), fmtEth(a.Expected))
	case RewardVerdictUnknown:
		amt := "?"
		if a.AmtArg != nil {
			amt = fmtEth(a.AmtArg) + " ETH"
		}
		log.Warnf("Reward at block %d by OC %s: %s (passed %s, paid %s ETH to %s): %s",
			a.Block, caller, RewardVerdictUnknown, amt, fmtEth(a.Paid), a.Winner.Hex(), a.Note)
	default:
		log.Printf("Reward at block %d by OC %s: %s (passed %s ETH = balance - fees owed; paid %s ETH to %s)",
			a.Block, caller, RewardVerdictOK, fmtEth(a.AmtArg), fmtEth(a.Paid), a.Winner.Hex())
	}
}

// checkEpochAdvance runs from the vote loop each cycle. When startBlock has
// moved since the previous cycle, some node rewarded the epoch: audit the
// reward(s) that landed since the old epoch end and log the fee reserve.
func checkEpochAdvance(cProps *ConnectionProps, startBlock *big.Int, interval uint16, head uint64) {
	prev := cProps.lastEpochStart
	cProps.lastEpochStart = new(big.Int).Set(startBlock)
	if prev == nil || startBlock.Cmp(prev) <= 0 {
		return
	}

	oldEnd := prev.Uint64() + uint64(interval)
	// head was read before startBlock; a reward mined in between sits one
	// block past it. Re-read so the audit window always covers the reward
	// that moved startBlock.
	if latest, err := cProps.Client.BlockNumber(context.Background()); err == nil && latest > head {
		head = latest
	}
	log.Infof("Epoch advanced (start block %s -> %s): auditing the reward that closed it", prev, startBlock)
	audits, err := AuditRewards(cProps, oldEnd, head)
	if err != nil {
		log.Warnf("Could not audit the reward that advanced the epoch: %v", err)
		return
	}
	if len(audits) == 0 {
		log.Warnf("Epoch advanced but no Rwd event was found in blocks %d-%d", oldEnd, head)
	}
	for _, a := range audits {
		logRewardAudit(a, cProps.MyPubKey)
	}
	reserve, err := ReadFeeReserve(cProps, nil)
	if err != nil {
		log.Warnf("Could not read the fee reserve: %v", err)
		return
	}
	if reserve.Underfunded() {
		log.Warnf("Fee reserve: %s", reserve)
	} else {
		log.Infof("Fee reserve: %s", reserve)
	}
}
