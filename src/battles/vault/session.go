package vault

// Session: every on-chain operation the battles CLI performs, over a Client
// that is either a real ethclient or a simulated backend in tests.

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	vaultabi "ktp2/src/abis/vault"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Client is what the session needs from an Ethereum node. ethclient.Client
// and simulated.Backend.Client() both satisfy it.
type Client interface {
	bind.ContractBackend
	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
}

// DefaultTxTimeout bounds every wait for a transaction to mine.
const DefaultTxTimeout = 5 * time.Minute

// DefaultScanChunk is the widest eth_getLogs range asked for at once. Free
// public RPC tiers reject anything wider; a refusal halves it and retries.
const DefaultScanChunk uint64 = 10_000

// minScanChunk is where halving stops and the error is surfaced instead.
const minScanChunk uint64 = 10

// Session binds one signer to one deployed Vault.
type Session struct {
	Client    Client
	Vault     *vaultabi.BurnBankVault
	VaultAddr common.Address
	From      common.Address
	TxTimeout time.Duration
	// DeployBlock is where event scans start: the block the Vault was
	// deployed in, from the config or found by NewSession.
	DeployBlock uint64
	// ScanChunk is the widest block range one log query asks for.
	ScanChunk uint64

	key     *ecdsa.PrivateKey
	chainID *big.Int
}

// NewSession connects cfg's signer to cfg's Vault. It verifies that the
// endpoint is on cfg's chain and that code is deployed at the address, and
// locates the deploy block when the config does not carry it.
func NewSession(ctx context.Context, client Client, cfg *Config) (*Session, error) {
	key, from, err := cfg.Signer()
	if err != nil {
		return nil, err
	}
	if cfg.VaultAddr == (common.Address{}) {
		return nil, fmt.Errorf("no vault address configured")
	}
	if cfg.ChainID == nil {
		return nil, fmt.Errorf("no chain ID configured")
	}
	if actual, err := client.ChainID(ctx); err != nil {
		return nil, fmt.Errorf("failed to read the endpoint's chain ID: %w", err)
	} else if actual.Cmp(cfg.ChainID) != 0 {
		return nil, fmt.Errorf("endpoint is on chain %s but %s says %s; fix %s or %s", actual, ConfigFileName, cfg.ChainID, EnvEndpoint, EnvChainID)
	}
	code, err := client.CodeAt(ctx, cfg.VaultAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check the vault address: %w", err)
	}
	if len(code) == 0 {
		return nil, fmt.Errorf("no contract at %s on chain %s", cfg.VaultAddr.Hex(), cfg.ChainID)
	}
	v, err := vaultabi.NewBurnBankVault(cfg.VaultAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind the vault: %w", err)
	}
	s := &Session{Client: client, Vault: v, VaultAddr: cfg.VaultAddr, From: from, TxTimeout: DefaultTxTimeout, DeployBlock: cfg.VaultBlock, ScanChunk: DefaultScanChunk, key: key, chainID: cfg.ChainID}
	if s.DeployBlock == 0 {
		if s.DeployBlock, err = findDeployBlock(ctx, client, cfg.VaultAddr); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// findDeployBlock binary-searches for the first block that holds code at
// addr: about 25 CodeAt calls on mainnet, done once per session when the
// config does not record the block.
func findDeployBlock(ctx context.Context, client Client, addr common.Address) (uint64, error) {
	head, err := client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to read the chain head: %w", err)
	}
	lo, hi := uint64(0), head
	for lo < hi {
		mid := lo + (hi-lo)/2
		code, err := client.CodeAt(ctx, addr, new(big.Int).SetUint64(mid))
		if err != nil {
			return 0, fmt.Errorf("failed to find the vault's deploy block (set %s to skip the search): %w", EnvVaultBlock, err)
		}
		if len(code) > 0 {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo, nil
}

// Deploy publishes a new Vault whitelisting banks and waits for it to mine.
func Deploy(ctx context.Context, client Client, key *ecdsa.PrivateKey, chainID *big.Int, banks []common.Address) (common.Address, *types.Receipt, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to build transactor: %w", err)
	}
	auth.Context = ctx
	if banks == nil {
		banks = []common.Address{}
	}
	addr, tx, _, err := vaultabi.DeployBurnBankVault(auth, client, banks)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("deploy failed: %w", err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, DefaultTxTimeout)
	defer cancel()
	rcpt, err := bind.WaitMined(waitCtx, client, tx)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("deploy tx %s not mined: %w", tx.Hash().Hex(), err)
	}
	if rcpt.Status != types.ReceiptStatusSuccessful {
		return common.Address{}, rcpt, fmt.Errorf("deploy tx %s reverted", tx.Hash().Hex())
	}
	return addr, rcpt, nil
}

func (s *Session) auth(ctx context.Context) (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(s.key, s.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to build transactor: %w", err)
	}
	auth.Context = ctx
	return auth, nil
}

// send runs one contract transaction and waits for its receipt. A revert at
// gas estimation surfaces as the contract's reason string; a revert on-chain
// surfaces as an error naming the tx.
func (s *Session) send(ctx context.Context, label string, fn func(*bind.TransactOpts) (*types.Transaction, error)) (*types.Receipt, error) {
	auth, err := s.auth(ctx)
	if err != nil {
		return nil, err
	}
	tx, err := fn(auth)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	timeout := s.TxTimeout
	if timeout == 0 {
		timeout = DefaultTxTimeout
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	rcpt, err := bind.WaitMined(waitCtx, s.Client, tx)
	if err != nil {
		return nil, fmt.Errorf("%s: tx %s not mined within %s: %w", label, tx.Hash().Hex(), timeout, err)
	}
	if rcpt.Status != types.ReceiptStatusSuccessful {
		return rcpt, fmt.Errorf("%s: tx %s reverted", label, tx.Hash().Hex())
	}
	return rcpt, nil
}

// Fund locks amount wei under keyHash.
func (s *Session) Fund(ctx context.Context, keyHash common.Hash, amount *big.Int) (*types.Receipt, error) {
	return s.send(ctx, "fund", func(a *bind.TransactOpts) (*types.Transaction, error) {
		a.Value = amount
		return s.Vault.Fund(a, keyHash)
	})
}

// Unlock spends the vault for key into bank's give().
func (s *Session) Unlock(ctx context.Context, key Key, bank common.Address) (*types.Receipt, error) {
	return s.send(ctx, "unlock", func(a *bind.TransactOpts) (*types.Transaction, error) {
		return s.Vault.Unlock(a, key, bank)
	})
}

// AddBank whitelists a Burn Bank (owner only).
func (s *Session) AddBank(ctx context.Context, bank common.Address) (*types.Receipt, error) {
	return s.send(ctx, "addBank", func(a *bind.TransactOpts) (*types.Transaction, error) {
		return s.Vault.AddBank(a, bank)
	})
}

// RemoveBank un-whitelists a Burn Bank (owner only).
func (s *Session) RemoveBank(ctx context.Context, bank common.Address) (*types.Receipt, error) {
	return s.send(ctx, "removeBank", func(a *bind.TransactOpts) (*types.Transaction, error) {
		return s.Vault.RemoveBank(a, bank)
	})
}

func (s *Session) callOpts(ctx context.Context) *bind.CallOpts {
	return &bind.CallOpts{Context: ctx, From: s.From}
}

// scan runs fn over [DeployBlock, head] in ScanChunk-sized windows, halving
// the window when the provider refuses a range, so scans work on RPCs that
// cap eth_getLogs.
func (s *Session) scan(ctx context.Context, fromBlock uint64, fn func(opts *bind.FilterOpts) error) error {
	head, err := s.Client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to read the chain head: %w", err)
	}
	start := s.DeployBlock
	if fromBlock > start {
		start = fromBlock
	}
	chunk := s.ScanChunk
	if chunk == 0 {
		chunk = DefaultScanChunk
	}
	for start <= head {
		end := start + chunk - 1
		if end > head || end < start {
			end = head
		}
		endCopy := end
		err := fn(&bind.FilterOpts{Context: ctx, Start: start, End: &endCopy})
		if err != nil {
			if isLogLimitError(err) && chunk > minScanChunk {
				chunk /= 2
				continue
			}
			return err
		}
		if end == head {
			break
		}
		start = end + 1
	}
	return nil
}

// isLogLimitError recognizes the range refusals public RPCs return.
func isLogLimitError(err error) bool {
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{"log limit", "query returned more than", "block range", "blocks range", "range is too large", "exceeds", "too many"} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

// Banks lists the currently whitelisted Burn Banks, reconstructed from
// BankAdded events and confirmed against the banks mapping (so a removed
// bank drops out). Order is first-added first.
func (s *Session) Banks(ctx context.Context) ([]common.Address, error) {
	seen := make(map[common.Address]bool)
	var out []common.Address
	err := s.scan(ctx, 0, func(opts *bind.FilterOpts) error {
		iter, err := s.Vault.FilterBankAdded(opts, nil)
		if err != nil {
			return err
		}
		defer iter.Close()
		for iter.Next() {
			b := iter.Event.Bank
			if seen[b] {
				continue
			}
			seen[b] = true
			active, err := s.Vault.Banks(s.callOpts(ctx), b)
			if err != nil {
				return fmt.Errorf("failed to read banks(%s): %w", b.Hex(), err)
			}
			if active {
				out = append(out, b)
			}
		}
		return iter.Error()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan BankAdded events: %w", err)
	}
	return out, nil
}

// IsBank reports whether bank is whitelisted right now.
func (s *Session) IsBank(ctx context.Context, bank common.Address) (bool, error) {
	return s.Vault.Banks(s.callOpts(ctx), bank)
}

// VaultStatus is one vault as the contract sees it.
type VaultStatus struct {
	KeyHash     common.Hash
	Funder      common.Address
	Amount      *big.Int
	FundedBlock uint64
	Spent       bool
	Unlockable  bool
}

// Status reads the vault stored under keyHash.
func (s *Session) Status(ctx context.Context, keyHash common.Hash) (*VaultStatus, error) {
	v, err := s.Vault.Vaults(s.callOpts(ctx), keyHash)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault %s: %w", keyHash.Hex(), err)
	}
	return &VaultStatus{
		KeyHash:     keyHash,
		Funder:      v.Funder,
		Amount:      v.Amount,
		FundedBlock: v.FundedBlock.Uint64(),
		Spent:       v.Spent,
		Unlockable:  v.Amount.Sign() > 0 && !v.Spent,
	}, nil
}

// Challenge is a vault's full history: its funding and, if spent, its unlock.
type Challenge struct {
	KeyHash     common.Hash
	Funder      common.Address
	Amount      *big.Int
	FundedBlock uint64
	Spent       bool
	Bank        common.Address // zero until unlocked
	Unlocker    common.Address // zero until unlocked
	UnlockBlock uint64         // zero until unlocked
}

// Challenges lists every vault funded at or after fromBlock, oldest first.
func (s *Session) Challenges(ctx context.Context, fromBlock uint64) ([]Challenge, error) {
	byHash := make(map[common.Hash]*Challenge)
	var out []*Challenge
	err := s.scan(ctx, fromBlock, func(opts *bind.FilterOpts) error {
		funded, err := s.Vault.FilterVaultFunded(opts, nil, nil)
		if err != nil {
			return err
		}
		defer funded.Close()
		for funded.Next() {
			e := funded.Event
			c := &Challenge{KeyHash: e.KeyHash, Funder: e.Funder, Amount: e.Amount, FundedBlock: e.Raw.BlockNumber}
			byHash[e.KeyHash] = c
			out = append(out, c)
		}
		if err := funded.Error(); err != nil {
			return err
		}
		unlocked, err := s.Vault.FilterVaultUnlocked(opts, nil, nil, nil)
		if err != nil {
			return err
		}
		defer unlocked.Close()
		for unlocked.Next() {
			e := unlocked.Event
			if c, ok := byHash[e.KeyHash]; ok {
				c.Spent, c.Bank, c.Unlocker, c.UnlockBlock = true, e.Bank, e.Unlocker, e.Raw.BlockNumber
			}
		}
		return unlocked.Error()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan vault events: %w", err)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].FundedBlock < out[j].FundedBlock })
	result := make([]Challenge, len(out))
	for i, c := range out {
		result[i] = *c
	}
	return result, nil
}

// SenderOf recovers the signer of a mined transaction.
func SenderOf(tx *types.Transaction) (common.Address, error) {
	return types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
}

// AddressOf is the address a private key signs as.
func AddressOf(key *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(key.PublicKey)
}
