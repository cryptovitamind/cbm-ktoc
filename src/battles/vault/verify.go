package vault

// Match verification: the one off-chain judgment in the challenge flow.
// A participant claims a transaction fulfilled a challenge; the site checks
// that it was a successful give() call to a whitelisted Burn Bank, for at
// least the challenge amount, mined after the vault was funded.

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Reasons a match is rejected; the CLI prints them and tests assert on them.
const (
	MatchOK              = "ok"
	MatchNotMined        = "transaction is not mined"
	MatchReverted        = "transaction reverted"
	MatchNotABank        = "recipient is not a whitelisted Burn Bank"
	MatchNotGive         = "transaction is not a give() call"
	MatchTooSmall        = "amount is below the challenge minimum"
	MatchTooEarly        = "transaction was mined before the vault was funded"
	MatchMissingContract = "transaction has no recipient"
)

// giveSelector is the 4-byte selector of Ktv2.give(); the test pins it
// against the compiled ABI.
var giveSelector = []byte{0, 0, 0, 0}

// MatchResult is the verdict on one claimed match.
type MatchResult struct {
	OK     bool
	Reason string
	Bank   common.Address
	Sender common.Address
	Amount *big.Int
	Block  uint64
}

// EvaluateMatch is the pure rule: given the transaction, its receipt, whether
// the recipient is a whitelisted bank, and the challenge terms, decide.
func EvaluateMatch(tx *types.Transaction, receipt *types.Receipt, sender common.Address, isBank bool, minAmount *big.Int, fundedBlock uint64) MatchResult {
	return MatchResult{Reason: "not implemented"}
}

// VerifyMatch loads the transaction and receipt, checks the whitelist, and
// evaluates it against the challenge funded under keyHash.
func (s *Session) VerifyMatch(ctx context.Context, txHash common.Hash, keyHash common.Hash) (*MatchResult, error) {
	return nil, errNotImplemented
}
