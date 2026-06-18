package ktfunc

import (
	"context"
	"crypto/ecdsa"
	"ktp2/src/abis/ktv2"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	DefaultGasLimit     uint64 = 24000
	DefaultBlocksToWait uint64 = 10
	DefaultChunkSize    int    = 500

	// SeedOffset is the fixed number of blocks past the epoch end at which
	// the lottery seed is sampled: seedBlock = endBlock + SeedOffset, and
	// the winner is selected from that block's hash. 32 (~6 min on mainnet)
	// keeps the seed block deep enough that it has effectively settled before
	// any operator reads it, so every node sees the same hash.
	//
	// CONSENSUS-CRITICAL: this MUST be identical across every operator. It
	// is a compile-time constant, NOT a flag or env var, precisely so two
	// nodes can never seed the lottery from different blocks and disagree on
	// the winner. The seed location must never be made operator-configurable.
	SeedOffset uint64 = 32

	// DefaultConfirmationDepth is how many blocks past the seed block the
	// node waits before SUBMITTING its vote, so the seed block is buried
	// under enough confirmations to have settled. It controls submission
	// timing / reorg burial ONLY — it does NOT change which block seeds the
	// lottery (that is always endBlock + SeedOffset). Safe to differ between
	// operators. 5 ≈ 1 min on a 12s/block chain.
	DefaultConfirmationDepth uint64 = 5

	// reorgSafetyDepth is how far behind the chain head a cached tip must be
	// before we record its block hash for reorg detection. Tips within this
	// many blocks of head are still subject to ordinary PoS reorgs and can
	// read inconsistently across load-balanced RPC backends, so recording
	// their hash would produce false-positive reorg wipes. 32 sits well past
	// typical reorg depth while staying below 2-epoch finality (~64).
	reorgSafetyDepth uint64 = 32
)

// TimeToWaitForBlocks is the polling interval used by the WaitForBlocks
// fallback path. Declared as `var` (not `const`) so tests can override
// it to keep their runtime short.
var TimeToWaitForBlocks time.Duration = 5 * time.Second

// DefaultWaitDuration is the pause between vote-cycle iterations of the
// continuous (`-run`) loop. Distinct from TimeToWaitForBlocks (the inner
// block-polling interval). One minute keeps idle RPC volume low.
var DefaultWaitDuration time.Duration = 60 * time.Second

type EthClient interface {
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	BlockNumber(ctx context.Context) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
	// SubscribeNewHead delivers new block headers as the chain advances.
	// On HTTP-only endpoints this typically returns an error — callers
	// should fall back to polling BlockNumber.
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
}

type StakedIterator interface {
	Next() bool
	Event() *ktv2.Ktv2Staked
	Error() error
	Close() error
}

type WithdrewIterator interface {
	Next() bool
	Event() *ktv2.Ktv2Withdrew
	Error() error
	Close() error
}

type VotedIterator interface {
	Next() bool
	Event() *ktv2.Ktv2Voted
	Error() error
	Close() error
}

type RwdIterator interface {
	Next() bool
	Event() *ktv2.Ktv2Rwd
	Error() error
	Close() error
}

type Ktv2Interface interface {
	StartBlock(opts *bind.CallOpts) (*big.Int, error)
	EpochInterval(opts *bind.CallOpts) (uint16, error)
	UserStks(opts *bind.CallOpts, address common.Address) (*big.Int, error)
	Vote(opts *bind.TransactOpts, recipient common.Address, data string) (*types.Transaction, error)
	ResetVote(opts *bind.TransactOpts, recipient common.Address) (*types.Transaction, error)
	Rwd(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error)
	BlockRwd(opts *bind.CallOpts, blockNumber *big.Int, recipient common.Address) (uint16, error)
	ConsensusReq(opts *bind.CallOpts) (uint16, error)
	TotalOC(opts *bind.CallOpts) (uint16, error)
	FilterStaked(opts *bind.FilterOpts) (StakedIterator, error)
	FilterWithdrew(opts *bind.FilterOpts) (WithdrewIterator, error)
	Give(opts *bind.TransactOpts) (*types.Transaction, error)
	WithdrawOCFee(opts *bind.TransactOpts) (*types.Transaction, error)
	PastOcFees(opts *bind.CallOpts, oc common.Address) (*big.Int, error)
	VoteToAdd(opts *bind.TransactOpts, newOC common.Address, data string) (*types.Transaction, error)
	VoteToRemove(opts *bind.TransactOpts, existingOC common.Address, data string) (*types.Transaction, error)
	ResetVoteToAdd(opts *bind.TransactOpts, newOC common.Address) (*types.Transaction, error)
	ResetVoteToRemove(opts *bind.TransactOpts, existingOC common.Address) (*types.Transaction, error)
	SetEpochInterval(opts *bind.TransactOpts, newInterval uint16) (*types.Transaction, error)
	SetOCFee(opts *bind.TransactOpts, fee uint16) (*types.Transaction, error)

	TotalStk(opts *bind.CallOpts) (*big.Int, error)
	TotalGvn(opts *bind.CallOpts) (*big.Int, error)
	TotalBurned(opts *bind.CallOpts) (*big.Int, error)
	MaxBrnPrc(opts *bind.CallOpts) (uint16, error)
	DonationPrc(opts *bind.CallOpts) (uint16, error)
	BurnFactor(opts *bind.CallOpts) (uint16, error)
	V2(opts *bind.CallOpts) (bool, error)
	OcFee(opts *bind.CallOpts) (uint16, error)
	OcFees(opts *bind.CallOpts, oc common.Address, blockNumber *big.Int) (*big.Int, error)
	TlOcFees(opts *bind.CallOpts) (*big.Int, error)
	Stake(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error)
	Withdraw(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error)

	FilterRwd(opts *bind.FilterOpts) (RwdIterator, error)
	FilterVoted(opts *bind.FilterOpts) (VotedIterator, error)

	OcRwdrs(opts *bind.CallOpts, address common.Address) (bool, error)
	Declines(opts *bind.CallOpts, address common.Address) (bool, error)
	HasVotedAdd(opts *bind.CallOpts, voter common.Address, target common.Address) (bool, error)
	HasVotedRemove(opts *bind.CallOpts, voter common.Address, target common.Address) (bool, error)
	Owner(opts *bind.CallOpts) (common.Address, error)
}

// ConnectionProps holds Ethereum connection properties and contract instances.
type ConnectionProps struct {
	ChainID        *big.Int             // Blockchain chain ID
	Client         EthClient            // Ethereum client connection
	Backend        bind.ContractBackend // Contract backend for KT contract
	MyPubKey       common.Address       // User's public address
	MyPrivateKey   *ecdsa.PrivateKey    // User's private key (for testing only)
	Addresses      *Addresses           // Contract and wallet addresses
	KtAddr         common.Address       // KT contract address
	KtBlock        *big.Int             // Start block number for KT contract
	Kt             Ktv2Interface        // KT contract instance
	GasLimit       uint64               // Gas limit for transactions
	BlocksToWait   uint64               // Number of blocks to wait for transactions to confirm
	QueryDelay     time.Duration        // Delay between API queries in milliseconds to prevent rate limiting
	V2Uniswap      bool                 // If true, use Uniswap V2, else V1.
	ChunkSize      int                  // Size of chunks for processing large data sets
	WaitDuration   time.Duration        // Duration to wait between operations
	// CacheDir is the directory for the on-disk event/fees caches. Empty means
	// the default "cache". Set a distinct dir per node when running several
	// operator instances on one machine so their bbolt caches don't collide.
	CacheDir string
	// DeclinesCache memoizes Declines() lookups for the lifetime of the
	// process. Declines is a contract state read and rarely changes, so
	// re-querying it every epoch is wasteful. Nil = first use will create it.
	DeclinesCache map[common.Address]bool

	// ConfirmationDepth — how many blocks past the SEED block (endBlock +
	// SeedOffset) the node waits before submitting its vote, so the seed
	// block is buried under enough confirmations to have settled. Controls
	// submission timing / reorg burial ONLY; it does NOT change which block
	// seeds the lottery. Zero means "use DefaultConfirmationDepth".
	ConfirmationDepth uint64

	// cachedGasPrice memoizes SuggestGasPrice with a short TTL (60s). Gas
	// price is cosmetic / a tx default only — never gates consensus or a
	// transaction's success — so a stale read is harmless. Concurrent-safe
	// (sync.Mutex inside cachedValue). Do not read or write directly.
	// Contract state that DOES gate a tx or the seed is intentionally never
	// cached here; see state_cache.go.
	cachedGasPrice cachedValue[*big.Int]
}

// ResolvedCacheDir returns the directory for on-disk caches, defaulting to
// "cache" when CacheDir is unset.
func (cProps *ConnectionProps) ResolvedCacheDir() string {
	if cProps.CacheDir != "" {
		return cProps.CacheDir
	}
	return "cache"
}

// Addresses holds Ethereum addresses and private keys from environment variables.
type Addresses struct {
	MyPublicKey  string // User's public key (hex string)
	MyPrivateKey string // User's private key (hex string, for testing only)
	DeadAddr     string // Dead address for burning tokens
	TargetAddr   string // Target address for operations
	FactoryAddr  string // Factory contract address
	PoolAddr     string // Pool address
	TknAddr      string // Token contract address
	TknPrcAddr   string // Token price contract address
	KtAddr       string // KT contract address
	EthEndpoint  string // Ethereum client endpoint
	KtStartBlock string // Start block number for KT contract
	WaitDuration string // Duration to wait between operations (e.g., "1s", "2m")
}

// UserStakeData holds staking-related data for a user.
type UserStakeData struct {
	StakeAmount *big.Int   // Amount of tokens staked
	Prob        *big.Float // Probability (likely for voting or rewards)
}

type StakedIteratorWrapper struct {
	*ktv2.Ktv2StakedIterator
}

func (w *StakedIteratorWrapper) Event() *ktv2.Ktv2Staked {
	return w.Ktv2StakedIterator.Event
}

func (w *StakedIteratorWrapper) Close() error {
	return w.Ktv2StakedIterator.Close()
}

type WithdrewIteratorWrapper struct {
	*ktv2.Ktv2WithdrewIterator
}

func (w *WithdrewIteratorWrapper) Event() *ktv2.Ktv2Withdrew {
	return w.Ktv2WithdrewIterator.Event
}

func (w *WithdrewIteratorWrapper) Close() error {
	return w.Ktv2WithdrewIterator.Close()
}

type Ktv2Wrapper struct {
	*ktv2.Ktv2
}

func (w *Ktv2Wrapper) FilterStaked(opts *bind.FilterOpts) (StakedIterator, error) {
	iter, err := w.Ktv2.FilterStaked(opts)
	if err != nil {
		return nil, err
	}
	return &StakedIteratorWrapper{iter}, nil
}

func (w *Ktv2Wrapper) FilterWithdrew(opts *bind.FilterOpts) (WithdrewIterator, error) {
	iter, err := w.Ktv2.FilterWithdrew(opts)
	if err != nil {
		return nil, err
	}
	return &WithdrewIteratorWrapper{iter}, nil
}

type VotedIteratorWrapper struct {
	*ktv2.Ktv2VotedIterator
}

func (w *VotedIteratorWrapper) Event() *ktv2.Ktv2Voted {
	return w.Ktv2VotedIterator.Event
}

func (w *VotedIteratorWrapper) Close() error {
	return w.Ktv2VotedIterator.Close()
}

type RwdIteratorWrapper struct {
	*ktv2.Ktv2RwdIterator
}

func (w *RwdIteratorWrapper) Event() *ktv2.Ktv2Rwd {
	return w.Ktv2RwdIterator.Event
}

func (w *RwdIteratorWrapper) Close() error {
	return w.Ktv2RwdIterator.Close()
}

func (w *Ktv2Wrapper) FilterVoted(opts *bind.FilterOpts) (VotedIterator, error) {
	iter, err := w.Ktv2.FilterVoted(opts)
	if err != nil {
		return nil, err
	}
	return &VotedIteratorWrapper{iter}, nil
}

func (w *Ktv2Wrapper) FilterRwd(opts *bind.FilterOpts) (RwdIterator, error) {
	iter, err := w.Ktv2.FilterRwd(opts)
	if err != nil {
		return nil, err
	}
	return &RwdIteratorWrapper{iter}, nil
}
