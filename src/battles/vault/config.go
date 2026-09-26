package vault

// battles.env: the CLI's configuration file. It holds a private key, so it is
// written 0600 and git-ignored. Every key is also honored as an environment
// variable of the same name, so a CI job can run without the file.

import (
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const (
	ConfigFileName = "battles.env"

	EnvEndpoint   = "BATTLES_ETH_ENDPOINT"
	EnvPrivateKey = "BATTLES_PRIVATE_KEY"
	EnvVaultAddr  = "BATTLES_VAULT_ADDR"
	EnvChainID    = "BATTLES_CHAIN_ID"
	EnvVaultBlock = "BATTLES_VAULT_BLOCK"

	configFileMode = 0o600
)

var errNotImplemented = errors.New("not implemented")

// Config is everything the CLI needs to talk to one Vault.
type Config struct {
	EthEndpoint string
	PrivateKey  string // hex, 0x optional
	VaultAddr   common.Address
	ChainID     *big.Int
	// VaultBlock is the block the Vault was deployed in: where event scans
	// start. Zero means unknown; the session then finds it by binary search.
	VaultBlock uint64
}

// LoadConfig reads path, then lets same-named environment variables override
// each value. With no file, the environment alone is used; when that is empty
// too, the error wraps fs.ErrNotExist so the caller can start the wizard.
func LoadConfig(path string) (*Config, error) {
	return nil, errNotImplemented
}

// Save writes the config to path with mode 0600.
func (c *Config) Save(path string) error {
	return errNotImplemented
}

// Validate checks that every field the CLI needs is present and well-formed.
func (c *Config) Validate() error {
	return errNotImplemented
}

// Signer parses the private key and returns it with its address.
func (c *Config) Signer() (*ecdsa.PrivateKey, common.Address, error) {
	return nil, common.Address{}, errNotImplemented
}
