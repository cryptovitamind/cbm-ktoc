package ktfunc

import (
	"context"
	"crypto/ecdsa"
	"ktp2/src/abis"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	DefaultGasLimit     uint64        = 24000
	DefaultBlocksToWait uint64        = 10
	TimeToWaitForBlocks time.Duration = 5 * time.Second
)

type EthClient interface {
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	BlockNumber(ctx context.Context) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type Ktv2Interface interface {
	StartBlock(opts *bind.CallOpts) (*big.Int, error)
	EpochInterval(opts *bind.CallOpts) (*big.Int, error)
	UserStks(opts *bind.CallOpts, address common.Address) (*big.Int, error)
	Vote(opts *bind.TransactOpts, recipient common.Address, data string) (*types.Transaction, error)
	Rwd(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error)
	BlockRwd(opts *bind.CallOpts, blockNumber *big.Int, recipient common.Address) (uint16, error)
	ConsensusReq(opts *bind.CallOpts) (uint16, error)
	FilterStaked(opts *bind.FilterOpts) (*abis.Ktv2StakedIterator, error)
	FilterWithdrew(opts *bind.FilterOpts) (*abis.Ktv2WithdrewIterator, error)
	Give(opts *bind.TransactOpts) (*types.Transaction, error)
	WithdrawOCFee(opts *bind.TransactOpts, blocks []uint32) (*types.Transaction, error)
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
	Kt           *abis.Ktv2           // KT contract instance
	GasLimit     uint64               // Gas limit for transactions
	BlocksToWait uint64               // Number of blocks to wait for transactions to confirm
}

// Addresses holds Ethereum addresses and private keys from environment variables.
type Addresses struct {
	MyPublicKey  string // User's public key (hex string)
	MyPrivateKey string // User's private key (hex string, for testing only)
	DeadAddr     string // Dead address for burning tokens
	TargetAddr   string // Target address for operations
	FactoryAddr  string // Factory contract address
	PoolAddr     string // Pool contract address
	TknAddr      string // Token contract address
	TknPrcAddr   string // Token price contract address
	KtAddr       string // KT contract address
	EthEndpoint  string // Ethereum client endpoint
	KtStartBlock string // Start block number for KT contract
}

// UserStakeData holds staking-related data for a user.
type UserStakeData struct {
	StakeAmount *big.Int   // Amount of tokens staked
	Prob        *big.Float // Probability (likely for voting or rewards)
}
