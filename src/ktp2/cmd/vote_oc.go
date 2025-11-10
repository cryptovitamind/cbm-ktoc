package main

import (
	"flag"
	"fmt"
	"ktp2/src/ktp2/ktfunc"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

var _ common.Address

// VoteFlags holds the voting-related flags.
type VoteFlags struct {
	VoteToRemoveOC      string
	VoteToAddOC         string
	ResetVoteToAddOC    string
	ResetVoteToRemoveOC string
	DataForOCVote       string
	PrintOCVoteEvents   string
}

// ParseVoteFlags parses the voting flags. This can be called from main.go's parseFlags if needed.
func ParseVoteFlags() VoteFlags {
	vf := VoteFlags{}
	flag.StringVar(&vf.VoteToRemoveOC, "voteToRemoveOC", "", "Vote to remove an OC with the given Ethereum address")
	flag.StringVar(&vf.VoteToAddOC, "voteToAddOC", "", "Vote to add an OC with the given Ethereum address")
	flag.StringVar(&vf.ResetVoteToAddOC, "resetVoteToAddOC", "", "Reset vote to add an OC with the given Ethereum address")
	flag.StringVar(&vf.ResetVoteToRemoveOC, "resetVoteToRemoveOC", "", "Reset vote to remove an OC with the given Ethereum address")
	flag.StringVar(&vf.DataForOCVote, "dataForOCVote", "", "Data string for the OC vote (required if using voting flags)")
	flag.StringVar(&vf.PrintOCVoteEvents, "printOCVoteEvents", "", "Print all OC vote events between blocks. Use this format for value: <fromBlock>:<toBlock>")
	return vf
}

// HandleVoting checks and performs voting or resetting if flags are set. Returns true if handled (to skip other ops).
func HandleVoting(cProps *ktfunc.ConnectionProps, vf VoteFlags) bool {
	if len(vf.PrintOCVoteEvents) > 0 {
		blocks := strings.Split(vf.PrintOCVoteEvents, ":")
		if len(blocks) != 2 {
			log.Fatal("Invalid format for -printOCVoteEvents. Use <fromBlock>:<toBlock>")
		}

		// Parse block numbers as *big.Int and check for conversion errors
		from, ok := big.NewInt(0).SetString(blocks[0], 10)
		if !ok {
			log.Fatalf("Invalid fromBlock value: %s", blocks[0])
		}
		to, ok := big.NewInt(0).SetString(blocks[1], 10)
		if !ok {
			log.Fatalf("Invalid toBlock value: %s", blocks[1])
		}

		ktfunc.LogOperationStart("Printing OC Vote Events")
		err := ktfunc.PrintVoteEvents(cProps, from, to)
		if err != nil {
			log.Fatalf("Printing vote events failed: %v", err)
		}
		log.Info("Vote events printed successfully")
		return true
	}

	voteFlagsSet := vf.VoteToRemoveOC != "" || vf.VoteToAddOC != ""
	resetFlagsSet := vf.ResetVoteToAddOC != "" || vf.ResetVoteToRemoveOC != ""

	if !voteFlagsSet && !resetFlagsSet {
		return false
	}

	if voteFlagsSet && resetFlagsSet {
		log.Fatal("Cannot mix voting and resetting flags")
	}

	if vf.VoteToRemoveOC != "" && vf.VoteToAddOC != "" {
		log.Fatal("Cannot use both -voteToRemoveOC and -voteToAddOC at the same time")
	}

	if vf.ResetVoteToAddOC != "" && vf.ResetVoteToRemoveOC != "" {
		log.Fatal("Cannot use both -resetVoteToAddOC and -resetVoteToRemoveOC at the same time")
	}

	// Determine the target address, operation type, and data
	var targetAddr common.Address
	var operation string
	var data string
	var opName string
	var err error

	if vf.VoteToAddOC != "" {
		targetAddr, err = ktfunc.ValidateAddress(vf.VoteToAddOC)
		if err != nil {
			log.Fatalf("Invalid address for vote to add OC: %v", err)
		}
		operation = "vote_add"
		data = vf.DataForOCVote
		opName = "Voting to add OC"
	} else if vf.VoteToRemoveOC != "" {
		targetAddr, err = ktfunc.ValidateAddress(vf.VoteToRemoveOC)
		if err != nil {
			log.Fatalf("Invalid address for vote to remove OC: %v", err)
		}
		operation = "vote_remove"
		data = vf.DataForOCVote
		opName = "Voting to remove OC"
	} else if vf.ResetVoteToAddOC != "" {
		targetAddr, err = ktfunc.ValidateAddress(vf.ResetVoteToAddOC)
		if err != nil {
			log.Fatalf("Invalid address for reset vote to add OC: %v", err)
		}
		operation = "reset_add"
		opName = "Resetting vote to add OC"
	} else if vf.ResetVoteToRemoveOC != "" {
		targetAddr, err = ktfunc.ValidateAddress(vf.ResetVoteToRemoveOC)
		if err != nil {
			log.Fatalf("Invalid address for reset vote to remove OC: %v", err)
		}
		operation = "reset_remove"
		opName = "Resetting vote to remove OC"
	}

	// Validate data for voting operations
	if operation == "vote_add" || operation == "vote_remove" {
		if data == "" {
			log.Warn("DataForOCVote was not supplied.")
		}
	}

	// Perform the specific operation
	ktfunc.LogOperationStart(opName)
	switch operation {
	case "vote_add":
		err = ktfunc.VoteToAdd(cProps, targetAddr, data)
	case "vote_remove":
		err = ktfunc.VoteToRemove(cProps, targetAddr, data)
	case "reset_add":
		err = ktfunc.ResetVoteToAdd(cProps, targetAddr)
	case "reset_remove":
		err = ktfunc.ResetVoteToRemove(cProps, targetAddr)
	}

	if err != nil {
		log.Fatalf("%s failed: %v", opName, err)
	}

	log.Info("Operation completed successfully")
	return true
}

// PrintOCUsage adds voting and reset flag descriptions to the help message.
func PrintOCUsage() {
	fmt.Fprintf(os.Stderr, "  -voteToRemoveOC <eth addr>      %s\n", "Vote to remove an OC with the given Ethereum address")
	fmt.Fprintf(os.Stderr, "  -voteToAddOC <eth addr>         %s\n", "Vote to add an OC with the given Ethereum address")
	fmt.Fprintf(os.Stderr, "  -resetVoteToAddOC <eth addr>    %s\n", "Reset vote to add an OC with the given Ethereum address")
	fmt.Fprintf(os.Stderr, "  -resetVoteToRemoveOC <eth addr> %s\n", "Reset vote to remove an OC with the given Ethereum address")
	fmt.Fprintf(os.Stderr, "  -dataForOCVote <string>         %s\n", "Data string for the OC vote (required if using voting flags)")
	fmt.Fprintf(os.Stderr, "  -printOCVoteEvents              %s\n", "Print all OC vote events between <fromBlock>:<toBlock>")
}
