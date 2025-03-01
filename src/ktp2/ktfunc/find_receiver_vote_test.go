package ktfunc

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
)

// Helper function to create a types.Block with a given block number
func createBlock(number int64) *types.Block {
	header := &types.Header{
		Number: big.NewInt(number),
	}
	return types.NewBlockWithHeader(header)
}

func TestIsTimeToVote_EndBlockNil(t *testing.T) {
	block := createBlock(100)
	result := IsTimeToVote(nil, block)
	if result {
		t.Errorf("Expected false when endBlock is nil, got true")
	}
}

func TestIsTimeToVote_BlockNil(t *testing.T) {
	endBlock := big.NewInt(100)
	result := IsTimeToVote(endBlock, nil)
	if result {
		t.Errorf("Expected false when block is nil, got true")
	}
}

func TestIsTimeToVote_BlockNumberNil(t *testing.T) {
	endBlock := big.NewInt(100)
	// Create a block with nil Number by using an empty header
	header := &types.Header{} // Number is nil by default
	block := types.NewBlockWithHeader(header)
	result := IsTimeToVote(endBlock, block)
	if result {
		t.Errorf("Expected false when block number is nil, got true")
	}
}

func TestIsTimeToVote_BeforeEndBlock(t *testing.T) {
	endBlock := big.NewInt(200)
	block := createBlock(100)
	result := IsTimeToVote(endBlock, block)
	if result {
		t.Errorf("Expected false when current block (%d) is before end block (%d), got true", block.Number().Uint64(), endBlock.Uint64())
	}
}

func TestIsTimeToVote_AtEndBlock(t *testing.T) {
	endBlock := big.NewInt(100)
	block := createBlock(100)
	result := IsTimeToVote(endBlock, block)
	if !result {
		t.Errorf("Expected true when current block (%d) equals end block (%d), got false", block.Number().Uint64(), endBlock.Uint64())
	}
}

func TestIsTimeToVote_AfterEndBlock(t *testing.T) {
	endBlock := big.NewInt(100)
	block := createBlock(150)
	result := IsTimeToVote(endBlock, block)
	if !result {
		t.Errorf("Expected true when current block (%d) is after end block (%d), got false", block.Number().Uint64(), endBlock.Uint64())
	}
}
