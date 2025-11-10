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
	DefaultGasLimit     uint64        = 24000
	DefaultBlocksToWait uint64        = 10
	TimeToWaitForBlocks time.Duration = 5 * time.Second
	DefaultChunkSize    int           = 500
)

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

type Ktv2Interface interface {
	StartBlock(opts *bind.CallOpts) (*big.Int, error)
	EpochInterval(opts *bind.CallOpts) (uint16, error)
	UserStks(opts *bind.CallOpts, address common.Address) (*big.Int, error)
	Vote(opts *bind.TransactOpts, recipient common.Address, data string) (*types.Transaction, error)
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

	FilterRwd(opts *bind.FilterOpts) (*ktv2.Ktv2RwdIterator, error)
	FilterVoted(opts *bind.FilterOpts) (*ktv2.Ktv2VotedIterator, error)

	OcRwdrs(opts *bind.CallOpts, address common.Address) (bool, error)
	HasVotedAdd(opts *bind.CallOpts, voter common.Address, target common.Address) (bool, error)
	HasVotedRemove(opts *bind.CallOpts, voter common.Address, target common.Address) (bool, error)
	Owner(opts *bind.CallOpts) (common.Address, error)
}

// ConnectionProps holds Ethereum connection properties and contract instances.
type ConnectionProps struct {
	ChainID      *big.Int             // Blockchain chain ID
	Client       EthClient            // Ethereum client connection
	Backend      bind.ContractBackend // Contract backend for KT contract
	MyPubKey     common.Address       // User's public address
	MyPrivateKey *ecdsa.PrivateKey    // User's private key (for testing only)
	Addresses    *Addresses           // Contract and wallet addresses
	KtAddr       common.Address       // KT contract address
	KtBlock      *big.Int             // Start block number for KT contract
	Kt           Ktv2Interface        // KT contract instance
	GasLimit     uint64               // Gas limit for transactions
	BlocksToWait uint64               // Number of blocks to wait for transactions to confirm
	QueryDelay   time.Duration        // Delay between API queries in milliseconds to prevent rate limiting
	V2Uniswap    bool                 // If true, use Uniswap V2, else V1.
	ChunkSize    int                  // Size of chunks for processing large data sets
	WaitDuration time.Duration        // Duration to wait between operations
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
