package vault

// Match verification: the one off-chain judgment in the challenge flow.
// A participant claims a transaction fulfilled a challenge; the site checks
// that it was a successful give() call to a whitelisted Burn Bank, for at
// least the challenge amount, mined after the vault was funded.

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
var giveSelector = crypto.Keccak256([]byte("give()"))[:4]

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
// the recipient is a whitelisted bank, and the challenge terms, decide. The
// funding block itself does not count as "after": a give() in the same block
// could have been mined ahead of the funding.
func EvaluateMatch(tx *types.Transaction, receipt *types.Receipt, sender common.Address, isBank bool, minAmount *big.Int, fundedBlock uint64) MatchResult {
	r := MatchResult{Sender: sender, Amount: tx.Value()}
	if tx.To() == nil {
		return reject(r, MatchMissingContract)
	}
	r.Bank = *tx.To()
	if receipt == nil || receipt.BlockNumber == nil {
		return reject(r, MatchNotMined)
	}
	r.Block = receipt.BlockNumber.Uint64()
	if receipt.Status != types.ReceiptStatusSuccessful {
		return reject(r, MatchReverted)
	}
	if !isBank {
		return reject(r, MatchNotABank)
	}
	if !bytes.Equal(tx.Data(), giveSelector) {
		return reject(r, MatchNotGive)
	}
	if tx.Value().Cmp(minAmount) < 0 {
		return reject(r, MatchTooSmall)
	}
	if r.Block <= fundedBlock {
		return reject(r, MatchTooEarly)
	}
	r.OK, r.Reason = true, MatchOK
	return r
}

func reject(r MatchResult, reason string) MatchResult {
	r.OK, r.Reason = false, reason
	return r
}

// VerifyMatch loads the transaction and receipt, checks the whitelist, and
// evaluates it against the challenge funded under keyHash.
func (s *Session) VerifyMatch(ctx context.Context, txHash common.Hash, keyHash common.Hash) (*MatchResult, error) {
	st, err := s.Status(ctx, keyHash)
	if err != nil {
		return nil, err
	}
	if st.Amount.Sign() == 0 {
		return nil, fmt.Errorf("no vault is funded under %s", keyHash.Hex())
	}
	tx, pending, err := s.Client.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to load transaction %s: %w", txHash.Hex(), err)
	}
	sender, err := SenderOf(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to recover the sender of %s: %w", txHash.Hex(), err)
	}
	var receipt *types.Receipt
	if !pending {
		receipt, err = s.Client.TransactionReceipt(ctx, txHash)
		if err != nil {
			return nil, fmt.Errorf("failed to load receipt for %s: %w", txHash.Hex(), err)
		}
	}
	isBank := false
	if tx.To() != nil {
		isBank, err = s.IsBank(ctx, *tx.To())
		if err != nil {
			return nil, err
		}
	}
	r := EvaluateMatch(tx, receipt, sender, isBank, st.Amount, st.FundedBlock)
	return &r, nil
}
