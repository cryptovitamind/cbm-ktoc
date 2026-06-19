package ktfunc

// Manual epoch-recovery operations. The automated lottery (VoteAndReward) can
// wedge an epoch if operators on mismatched builds cast divergent votes before
// the deterministic-seed fix: the already-cast votes are frozen for that epoch
// slot and the contract has no skip/abandon function. These let an operator
// deliberately converge a stuck epoch (undo their vote and/or force a vote for
// an agreed address) using the contract's own resetVote/vote functions.

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

// manualVoteData marks a Voted event that came from an operator override
// (-voteFor) rather than the deterministic lottery. It is intentionally not a
// block hash, so -verifyLastWinner surfaces it as a non-algorithmic vote.
const manualVoteData = "manual-override"

// VoteForAddress casts a vote for a specific address in the current epoch,
// bypassing the lottery to deliberately converge a stuck epoch on an agreed
// winner. The contract still enforces its rules: epoch complete, caller is an
// OC, target not declined, and caller hasn't already voted this epoch (use
// ResetLotteryVote first if they have). After voting it reports where consensus
// now stands for that candidate.
func VoteForAddress(cProps *ConnectionProps, recipient common.Address) error {
	log.Printf("Manually voting for %s (override of the automatic lottery)", recipient.Hex())
	if err := vote(cProps, recipient, manualVoteData); err != nil {
		return fmt.Errorf("manual vote for %s failed: %w", recipient.Hex(), err)
	}

	callOpts := &bind.CallOpts{Context: context.Background(), From: cProps.MyPubKey}
	startBlock, err := cProps.Kt.StartBlock(callOpts)
	if err != nil {
		log.Warnf("Voted, but failed to read start block to report status: %v", err)
		return nil
	}
	count, err := cProps.Kt.BlockRwd(callOpts, startBlock, recipient)
	if err != nil {
		log.Warnf("Voted, but failed to read vote count: %v", err)
		return nil
	}
	required, err := cProps.Kt.ConsensusReq(callOpts)
	if err != nil {
		log.Warnf("Voted, but failed to read consensus requirement: %v", err)
		return nil
	}
	log.Printf("Vote recorded for %s at epoch %d: %d/%d votes", recipient.Hex(), startBlock.Uint64(), count, required)
	if count >= required {
		log.Printf("Consensus reached for %s. It can now be rewarded (the running node will on its next cycle).", recipient.Hex())
	} else {
		log.Printf("Consensus not yet reached; %d more matching vote(s) needed.", required-count)
	}
	return nil
}

// ResetLotteryVote undoes this node's vote for the current epoch so it can
// re-cast (e.g. to converge a stuck epoch on an agreed winner). recipient must
// be the address this node previously voted for, since the contract decrements
// that candidate's tally. The contract requires the caller to have an active vote
// this epoch and the candidate to currently hold at least one vote.
func ResetLotteryVote(cProps *ConnectionProps, recipient common.Address) error {
	log.Printf("Resetting this node's vote for %s in the current epoch", recipient.Hex())
	auth, err := NewTransactor(cProps)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}
	tx, err := cProps.Kt.ResetVote(auth, recipient)
	if err != nil {
		return fmt.Errorf("failed to reset vote for %s: %w", recipient.Hex(), err)
	}
	log.Printf("Reset-vote transaction sent: %s", tx.Hash().Hex())
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for reset-vote transaction to be mined: %w", err)
	}
	log.Debugf("Reset-vote transaction mined in block: %d", receipt.BlockNumber.Uint64())
	if err := WaitForBlocks(cProps); err != nil {
		return fmt.Errorf("failed to wait for additional blocks: %w", err)
	}
	log.Printf("Vote reset. This node can vote again this epoch (e.g. -voteFor <addr>).")
	return nil
}
