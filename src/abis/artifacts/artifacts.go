// Package artifacts embeds the compiled ABI and creation bytecode of every
// contract under src/contracts, as written by src/contracts/compile.sh. Tests
// deploy real bytecode from here onto a simulated chain, and the battles CLI
// deploys the Vault from it, so a build never depends on a Solidity toolchain.
package artifacts

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

//go:embed artifacts.json
var raw []byte

// Contract names as they appear in artifacts.json.
const (
	BurnBankVault  = "BurnBankVault"
	Ktv2           = "Ktv2"
	Ktv2Factory    = "Ktv2Factory"
	MockERC20      = "MockERC20"
	StubTokenPrice = "StubTokenPrice"
)

// Artifact is one compiled contract.
type Artifact struct {
	ABI abi.ABI
	Bin []byte
}

type rawArtifact struct {
	ABI json.RawMessage `json:"abi"`
	Bin string          `json:"bin"`
}

var (
	once   sync.Once
	loaded map[string]rawArtifact
	errAll error
)

// Load returns the compiled contract with the given name.
func Load(name string) (*Artifact, error) {
	once.Do(func() { errAll = json.Unmarshal(raw, &loaded) })
	if errAll != nil {
		return nil, fmt.Errorf("artifacts.json is malformed: %w", errAll)
	}
	ra, ok := loaded[name]
	if !ok {
		return nil, fmt.Errorf("no artifact named %q (run src/contracts/compile.sh)", name)
	}
	parsed, err := abi.JSON(strings.NewReader(string(ra.ABI)))
	if err != nil {
		return nil, fmt.Errorf("artifact %s: bad ABI: %w", name, err)
	}
	bin, err := hex.DecodeString(strings.TrimPrefix(ra.Bin, "0x"))
	if err != nil {
		return nil, fmt.Errorf("artifact %s: bad bytecode: %w", name, err)
	}
	return &Artifact{ABI: parsed, Bin: bin}, nil
}
