// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ktv2

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// Ktv2MetaData contains all meta data concerning the Ktv2 contract.
var Ktv2MetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_burnDest\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_dest\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_pool\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_ocPrcAddr\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tp\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"_v2\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"Gave\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOC\",\"type\":\"address\"}],\"name\":\"NodeAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldOC\",\"type\":\"address\"}],\"name\":\"NodeRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"Rwd\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"Voted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOC\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"VotedToAdd\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"existingOC\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"VotedToRemove\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"Withdrew\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"addOCRwdr\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"addVotes\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"allow\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"blockRwd\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"burnDest\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"burnFactor\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"consensusReq\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decline\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"declines\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dest\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"donationPrc\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"epochInterval\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"give\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"hasVoted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"maxBrnPrc\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ocFee\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"ocFees\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"ocRwdrVote\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"ocRwdrs\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pool\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"removeOCRwdr\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"removeVotes\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"}],\"name\":\"resetVote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOC\",\"type\":\"address\"}],\"name\":\"resetVoteToAdd\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"existingOC\",\"type\":\"address\"}],\"name\":\"resetVoteToRemove\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"rwd\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"amt\",\"type\":\"uint16\"}],\"name\":\"setBurnFactor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"req\",\"type\":\"uint16\"}],\"name\":\"setConsensusReq\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setDest\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"amt\",\"type\":\"uint16\"}],\"name\":\"setDonationPrc\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"interval\",\"type\":\"uint16\"}],\"name\":\"setEpochInterval\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"amt\",\"type\":\"uint16\"}],\"name\":\"setMaxBurnPrc\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"fee\",\"type\":\"uint16\"}],\"name\":\"setOCFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_pool\",\"type\":\"address\"}],\"name\":\"setPool\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"_v2\",\"type\":\"bool\"}],\"name\":\"setV2\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amt\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"startBlock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tlOcFees\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tokenAddr\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalBurned\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalGvn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalOC\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalStk\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tp\",\"outputs\":[{\"internalType\":\"contractTPI\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"userStks\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"v2\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"vote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOC\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"voteToAdd\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"existingOC\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"voteToRemove\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amt\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32[]\",\"name\":\"blocks\",\"type\":\"uint32[]\"}],\"name\":\"withdrawOCFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"withdrawTkn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// Ktv2ABI is the input ABI used to generate the binding from.
// Deprecated: Use Ktv2MetaData.ABI instead.
var Ktv2ABI = Ktv2MetaData.ABI

// Ktv2 is an auto generated Go binding around an Ethereum contract.
type Ktv2 struct {
	Ktv2Caller     // Read-only binding to the contract
	Ktv2Transactor // Write-only binding to the contract
	Ktv2Filterer   // Log filterer for contract events
}

// Ktv2Caller is an auto generated read-only Go binding around an Ethereum contract.
type Ktv2Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Ktv2Transactor is an auto generated write-only Go binding around an Ethereum contract.
type Ktv2Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Ktv2Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Ktv2Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Ktv2Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Ktv2Session struct {
	Contract     *Ktv2             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Ktv2CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Ktv2CallerSession struct {
	Contract *Ktv2Caller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// Ktv2TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Ktv2TransactorSession struct {
	Contract     *Ktv2Transactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Ktv2Raw is an auto generated low-level Go binding around an Ethereum contract.
type Ktv2Raw struct {
	Contract *Ktv2 // Generic contract binding to access the raw methods on
}

// Ktv2CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Ktv2CallerRaw struct {
	Contract *Ktv2Caller // Generic read-only contract binding to access the raw methods on
}

// Ktv2TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Ktv2TransactorRaw struct {
	Contract *Ktv2Transactor // Generic write-only contract binding to access the raw methods on
}

// NewKtv2 creates a new instance of Ktv2, bound to a specific deployed contract.
func NewKtv2(address common.Address, backend bind.ContractBackend) (*Ktv2, error) {
	contract, err := bindKtv2(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ktv2{Ktv2Caller: Ktv2Caller{contract: contract}, Ktv2Transactor: Ktv2Transactor{contract: contract}, Ktv2Filterer: Ktv2Filterer{contract: contract}}, nil
}

// NewKtv2Caller creates a new read-only instance of Ktv2, bound to a specific deployed contract.
func NewKtv2Caller(address common.Address, caller bind.ContractCaller) (*Ktv2Caller, error) {
	contract, err := bindKtv2(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Ktv2Caller{contract: contract}, nil
}

// NewKtv2Transactor creates a new write-only instance of Ktv2, bound to a specific deployed contract.
func NewKtv2Transactor(address common.Address, transactor bind.ContractTransactor) (*Ktv2Transactor, error) {
	contract, err := bindKtv2(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Ktv2Transactor{contract: contract}, nil
}

// NewKtv2Filterer creates a new log filterer instance of Ktv2, bound to a specific deployed contract.
func NewKtv2Filterer(address common.Address, filterer bind.ContractFilterer) (*Ktv2Filterer, error) {
	contract, err := bindKtv2(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Ktv2Filterer{contract: contract}, nil
}

// bindKtv2 binds a generic wrapper to an already deployed contract.
func bindKtv2(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := Ktv2MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ktv2 *Ktv2Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ktv2.Contract.Ktv2Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ktv2 *Ktv2Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2.Contract.Ktv2Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ktv2 *Ktv2Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ktv2.Contract.Ktv2Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ktv2 *Ktv2CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ktv2.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ktv2 *Ktv2TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ktv2 *Ktv2TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ktv2.Contract.contract.Transact(opts, method, params...)
}

// AddVotes is a free data retrieval call binding the contract method 0xc0d4e0e0.
//
// Solidity: function addVotes(address ) view returns(uint16)
func (_Ktv2 *Ktv2Caller) AddVotes(opts *bind.CallOpts, arg0 common.Address) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "addVotes", arg0)

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// AddVotes is a free data retrieval call binding the contract method 0xc0d4e0e0.
//
// Solidity: function addVotes(address ) view returns(uint16)
func (_Ktv2 *Ktv2Session) AddVotes(arg0 common.Address) (uint16, error) {
	return _Ktv2.Contract.AddVotes(&_Ktv2.CallOpts, arg0)
}

// AddVotes is a free data retrieval call binding the contract method 0xc0d4e0e0.
//
// Solidity: function addVotes(address ) view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) AddVotes(arg0 common.Address) (uint16, error) {
	return _Ktv2.Contract.AddVotes(&_Ktv2.CallOpts, arg0)
}

// BlockRwd is a free data retrieval call binding the contract method 0xca1d1b87.
//
// Solidity: function blockRwd(uint256 , address ) view returns(uint16)
func (_Ktv2 *Ktv2Caller) BlockRwd(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "blockRwd", arg0, arg1)

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// BlockRwd is a free data retrieval call binding the contract method 0xca1d1b87.
//
// Solidity: function blockRwd(uint256 , address ) view returns(uint16)
func (_Ktv2 *Ktv2Session) BlockRwd(arg0 *big.Int, arg1 common.Address) (uint16, error) {
	return _Ktv2.Contract.BlockRwd(&_Ktv2.CallOpts, arg0, arg1)
}

// BlockRwd is a free data retrieval call binding the contract method 0xca1d1b87.
//
// Solidity: function blockRwd(uint256 , address ) view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) BlockRwd(arg0 *big.Int, arg1 common.Address) (uint16, error) {
	return _Ktv2.Contract.BlockRwd(&_Ktv2.CallOpts, arg0, arg1)
}

// BurnDest is a free data retrieval call binding the contract method 0xcb8cbb6e.
//
// Solidity: function burnDest() view returns(address)
func (_Ktv2 *Ktv2Caller) BurnDest(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "burnDest")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BurnDest is a free data retrieval call binding the contract method 0xcb8cbb6e.
//
// Solidity: function burnDest() view returns(address)
func (_Ktv2 *Ktv2Session) BurnDest() (common.Address, error) {
	return _Ktv2.Contract.BurnDest(&_Ktv2.CallOpts)
}

// BurnDest is a free data retrieval call binding the contract method 0xcb8cbb6e.
//
// Solidity: function burnDest() view returns(address)
func (_Ktv2 *Ktv2CallerSession) BurnDest() (common.Address, error) {
	return _Ktv2.Contract.BurnDest(&_Ktv2.CallOpts)
}

// BurnFactor is a free data retrieval call binding the contract method 0xb1b71afa.
//
// Solidity: function burnFactor() view returns(uint16)
func (_Ktv2 *Ktv2Caller) BurnFactor(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "burnFactor")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// BurnFactor is a free data retrieval call binding the contract method 0xb1b71afa.
//
// Solidity: function burnFactor() view returns(uint16)
func (_Ktv2 *Ktv2Session) BurnFactor() (uint16, error) {
	return _Ktv2.Contract.BurnFactor(&_Ktv2.CallOpts)
}

// BurnFactor is a free data retrieval call binding the contract method 0xb1b71afa.
//
// Solidity: function burnFactor() view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) BurnFactor() (uint16, error) {
	return _Ktv2.Contract.BurnFactor(&_Ktv2.CallOpts)
}

// ConsensusReq is a free data retrieval call binding the contract method 0xb34913eb.
//
// Solidity: function consensusReq() view returns(uint16)
func (_Ktv2 *Ktv2Caller) ConsensusReq(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "consensusReq")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// ConsensusReq is a free data retrieval call binding the contract method 0xb34913eb.
//
// Solidity: function consensusReq() view returns(uint16)
func (_Ktv2 *Ktv2Session) ConsensusReq() (uint16, error) {
	return _Ktv2.Contract.ConsensusReq(&_Ktv2.CallOpts)
}

// ConsensusReq is a free data retrieval call binding the contract method 0xb34913eb.
//
// Solidity: function consensusReq() view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) ConsensusReq() (uint16, error) {
	return _Ktv2.Contract.ConsensusReq(&_Ktv2.CallOpts)
}

// Declines is a free data retrieval call binding the contract method 0xde96e6c6.
//
// Solidity: function declines(address ) view returns(bool)
func (_Ktv2 *Ktv2Caller) Declines(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "declines", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Declines is a free data retrieval call binding the contract method 0xde96e6c6.
//
// Solidity: function declines(address ) view returns(bool)
func (_Ktv2 *Ktv2Session) Declines(arg0 common.Address) (bool, error) {
	return _Ktv2.Contract.Declines(&_Ktv2.CallOpts, arg0)
}

// Declines is a free data retrieval call binding the contract method 0xde96e6c6.
//
// Solidity: function declines(address ) view returns(bool)
func (_Ktv2 *Ktv2CallerSession) Declines(arg0 common.Address) (bool, error) {
	return _Ktv2.Contract.Declines(&_Ktv2.CallOpts, arg0)
}

// Dest is a free data retrieval call binding the contract method 0x84b366dc.
//
// Solidity: function dest() view returns(address)
func (_Ktv2 *Ktv2Caller) Dest(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "dest")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Dest is a free data retrieval call binding the contract method 0x84b366dc.
//
// Solidity: function dest() view returns(address)
func (_Ktv2 *Ktv2Session) Dest() (common.Address, error) {
	return _Ktv2.Contract.Dest(&_Ktv2.CallOpts)
}

// Dest is a free data retrieval call binding the contract method 0x84b366dc.
//
// Solidity: function dest() view returns(address)
func (_Ktv2 *Ktv2CallerSession) Dest() (common.Address, error) {
	return _Ktv2.Contract.Dest(&_Ktv2.CallOpts)
}

// DonationPrc is a free data retrieval call binding the contract method 0x135078ef.
//
// Solidity: function donationPrc() view returns(uint16)
func (_Ktv2 *Ktv2Caller) DonationPrc(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "donationPrc")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// DonationPrc is a free data retrieval call binding the contract method 0x135078ef.
//
// Solidity: function donationPrc() view returns(uint16)
func (_Ktv2 *Ktv2Session) DonationPrc() (uint16, error) {
	return _Ktv2.Contract.DonationPrc(&_Ktv2.CallOpts)
}

// DonationPrc is a free data retrieval call binding the contract method 0x135078ef.
//
// Solidity: function donationPrc() view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) DonationPrc() (uint16, error) {
	return _Ktv2.Contract.DonationPrc(&_Ktv2.CallOpts)
}

// EpochInterval is a free data retrieval call binding the contract method 0x09b1ef26.
//
// Solidity: function epochInterval() view returns(uint16)
func (_Ktv2 *Ktv2Caller) EpochInterval(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "epochInterval")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// EpochInterval is a free data retrieval call binding the contract method 0x09b1ef26.
//
// Solidity: function epochInterval() view returns(uint16)
func (_Ktv2 *Ktv2Session) EpochInterval() (uint16, error) {
	return _Ktv2.Contract.EpochInterval(&_Ktv2.CallOpts)
}

// EpochInterval is a free data retrieval call binding the contract method 0x09b1ef26.
//
// Solidity: function epochInterval() view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) EpochInterval() (uint16, error) {
	return _Ktv2.Contract.EpochInterval(&_Ktv2.CallOpts)
}

// HasVoted is a free data retrieval call binding the contract method 0x4d4d2b1c.
//
// Solidity: function hasVoted(address , address ) view returns(bool)
func (_Ktv2 *Ktv2Caller) HasVoted(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (bool, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "hasVoted", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasVoted is a free data retrieval call binding the contract method 0x4d4d2b1c.
//
// Solidity: function hasVoted(address , address ) view returns(bool)
func (_Ktv2 *Ktv2Session) HasVoted(arg0 common.Address, arg1 common.Address) (bool, error) {
	return _Ktv2.Contract.HasVoted(&_Ktv2.CallOpts, arg0, arg1)
}

// HasVoted is a free data retrieval call binding the contract method 0x4d4d2b1c.
//
// Solidity: function hasVoted(address , address ) view returns(bool)
func (_Ktv2 *Ktv2CallerSession) HasVoted(arg0 common.Address, arg1 common.Address) (bool, error) {
	return _Ktv2.Contract.HasVoted(&_Ktv2.CallOpts, arg0, arg1)
}

// MaxBrnPrc is a free data retrieval call binding the contract method 0x1a9f16bb.
//
// Solidity: function maxBrnPrc() view returns(uint16)
func (_Ktv2 *Ktv2Caller) MaxBrnPrc(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "maxBrnPrc")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// MaxBrnPrc is a free data retrieval call binding the contract method 0x1a9f16bb.
//
// Solidity: function maxBrnPrc() view returns(uint16)
func (_Ktv2 *Ktv2Session) MaxBrnPrc() (uint16, error) {
	return _Ktv2.Contract.MaxBrnPrc(&_Ktv2.CallOpts)
}

// MaxBrnPrc is a free data retrieval call binding the contract method 0x1a9f16bb.
//
// Solidity: function maxBrnPrc() view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) MaxBrnPrc() (uint16, error) {
	return _Ktv2.Contract.MaxBrnPrc(&_Ktv2.CallOpts)
}

// OcFee is a free data retrieval call binding the contract method 0xa64d755c.
//
// Solidity: function ocFee() view returns(uint16)
func (_Ktv2 *Ktv2Caller) OcFee(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "ocFee")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// OcFee is a free data retrieval call binding the contract method 0xa64d755c.
//
// Solidity: function ocFee() view returns(uint16)
func (_Ktv2 *Ktv2Session) OcFee() (uint16, error) {
	return _Ktv2.Contract.OcFee(&_Ktv2.CallOpts)
}

// OcFee is a free data retrieval call binding the contract method 0xa64d755c.
//
// Solidity: function ocFee() view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) OcFee() (uint16, error) {
	return _Ktv2.Contract.OcFee(&_Ktv2.CallOpts)
}

// OcFees is a free data retrieval call binding the contract method 0xb079861e.
//
// Solidity: function ocFees(address , uint256 ) view returns(uint256)
func (_Ktv2 *Ktv2Caller) OcFees(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "ocFees", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OcFees is a free data retrieval call binding the contract method 0xb079861e.
//
// Solidity: function ocFees(address , uint256 ) view returns(uint256)
func (_Ktv2 *Ktv2Session) OcFees(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Ktv2.Contract.OcFees(&_Ktv2.CallOpts, arg0, arg1)
}

// OcFees is a free data retrieval call binding the contract method 0xb079861e.
//
// Solidity: function ocFees(address , uint256 ) view returns(uint256)
func (_Ktv2 *Ktv2CallerSession) OcFees(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Ktv2.Contract.OcFees(&_Ktv2.CallOpts, arg0, arg1)
}

// OcRwdrVote is a free data retrieval call binding the contract method 0x3b98548f.
//
// Solidity: function ocRwdrVote(address , uint256 ) view returns(address)
func (_Ktv2 *Ktv2Caller) OcRwdrVote(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "ocRwdrVote", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OcRwdrVote is a free data retrieval call binding the contract method 0x3b98548f.
//
// Solidity: function ocRwdrVote(address , uint256 ) view returns(address)
func (_Ktv2 *Ktv2Session) OcRwdrVote(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _Ktv2.Contract.OcRwdrVote(&_Ktv2.CallOpts, arg0, arg1)
}

// OcRwdrVote is a free data retrieval call binding the contract method 0x3b98548f.
//
// Solidity: function ocRwdrVote(address , uint256 ) view returns(address)
func (_Ktv2 *Ktv2CallerSession) OcRwdrVote(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _Ktv2.Contract.OcRwdrVote(&_Ktv2.CallOpts, arg0, arg1)
}

// OcRwdrs is a free data retrieval call binding the contract method 0x6c6cdea8.
//
// Solidity: function ocRwdrs(address ) view returns(bool)
func (_Ktv2 *Ktv2Caller) OcRwdrs(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "ocRwdrs", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// OcRwdrs is a free data retrieval call binding the contract method 0x6c6cdea8.
//
// Solidity: function ocRwdrs(address ) view returns(bool)
func (_Ktv2 *Ktv2Session) OcRwdrs(arg0 common.Address) (bool, error) {
	return _Ktv2.Contract.OcRwdrs(&_Ktv2.CallOpts, arg0)
}

// OcRwdrs is a free data retrieval call binding the contract method 0x6c6cdea8.
//
// Solidity: function ocRwdrs(address ) view returns(bool)
func (_Ktv2 *Ktv2CallerSession) OcRwdrs(arg0 common.Address) (bool, error) {
	return _Ktv2.Contract.OcRwdrs(&_Ktv2.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ktv2 *Ktv2Caller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ktv2 *Ktv2Session) Owner() (common.Address, error) {
	return _Ktv2.Contract.Owner(&_Ktv2.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ktv2 *Ktv2CallerSession) Owner() (common.Address, error) {
	return _Ktv2.Contract.Owner(&_Ktv2.CallOpts)
}

// Pool is a free data retrieval call binding the contract method 0x16f0115b.
//
// Solidity: function pool() view returns(address)
func (_Ktv2 *Ktv2Caller) Pool(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "pool")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Pool is a free data retrieval call binding the contract method 0x16f0115b.
//
// Solidity: function pool() view returns(address)
func (_Ktv2 *Ktv2Session) Pool() (common.Address, error) {
	return _Ktv2.Contract.Pool(&_Ktv2.CallOpts)
}

// Pool is a free data retrieval call binding the contract method 0x16f0115b.
//
// Solidity: function pool() view returns(address)
func (_Ktv2 *Ktv2CallerSession) Pool() (common.Address, error) {
	return _Ktv2.Contract.Pool(&_Ktv2.CallOpts)
}

// RemoveVotes is a free data retrieval call binding the contract method 0xa2ca6dce.
//
// Solidity: function removeVotes(address ) view returns(uint16)
func (_Ktv2 *Ktv2Caller) RemoveVotes(opts *bind.CallOpts, arg0 common.Address) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "removeVotes", arg0)

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// RemoveVotes is a free data retrieval call binding the contract method 0xa2ca6dce.
//
// Solidity: function removeVotes(address ) view returns(uint16)
func (_Ktv2 *Ktv2Session) RemoveVotes(arg0 common.Address) (uint16, error) {
	return _Ktv2.Contract.RemoveVotes(&_Ktv2.CallOpts, arg0)
}

// RemoveVotes is a free data retrieval call binding the contract method 0xa2ca6dce.
//
// Solidity: function removeVotes(address ) view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) RemoveVotes(arg0 common.Address) (uint16, error) {
	return _Ktv2.Contract.RemoveVotes(&_Ktv2.CallOpts, arg0)
}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_Ktv2 *Ktv2Caller) StartBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "startBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_Ktv2 *Ktv2Session) StartBlock() (*big.Int, error) {
	return _Ktv2.Contract.StartBlock(&_Ktv2.CallOpts)
}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_Ktv2 *Ktv2CallerSession) StartBlock() (*big.Int, error) {
	return _Ktv2.Contract.StartBlock(&_Ktv2.CallOpts)
}

// TlOcFees is a free data retrieval call binding the contract method 0xdc09deaf.
//
// Solidity: function tlOcFees() view returns(uint256)
func (_Ktv2 *Ktv2Caller) TlOcFees(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "tlOcFees")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TlOcFees is a free data retrieval call binding the contract method 0xdc09deaf.
//
// Solidity: function tlOcFees() view returns(uint256)
func (_Ktv2 *Ktv2Session) TlOcFees() (*big.Int, error) {
	return _Ktv2.Contract.TlOcFees(&_Ktv2.CallOpts)
}

// TlOcFees is a free data retrieval call binding the contract method 0xdc09deaf.
//
// Solidity: function tlOcFees() view returns(uint256)
func (_Ktv2 *Ktv2CallerSession) TlOcFees() (*big.Int, error) {
	return _Ktv2.Contract.TlOcFees(&_Ktv2.CallOpts)
}

// TokenAddr is a free data retrieval call binding the contract method 0x5fbe4d1d.
//
// Solidity: function tokenAddr() view returns(address)
func (_Ktv2 *Ktv2Caller) TokenAddr(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "tokenAddr")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenAddr is a free data retrieval call binding the contract method 0x5fbe4d1d.
//
// Solidity: function tokenAddr() view returns(address)
func (_Ktv2 *Ktv2Session) TokenAddr() (common.Address, error) {
	return _Ktv2.Contract.TokenAddr(&_Ktv2.CallOpts)
}

// TokenAddr is a free data retrieval call binding the contract method 0x5fbe4d1d.
//
// Solidity: function tokenAddr() view returns(address)
func (_Ktv2 *Ktv2CallerSession) TokenAddr() (common.Address, error) {
	return _Ktv2.Contract.TokenAddr(&_Ktv2.CallOpts)
}

// TotalBurned is a free data retrieval call binding the contract method 0xd89135cd.
//
// Solidity: function totalBurned() view returns(uint256)
func (_Ktv2 *Ktv2Caller) TotalBurned(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "totalBurned")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalBurned is a free data retrieval call binding the contract method 0xd89135cd.
//
// Solidity: function totalBurned() view returns(uint256)
func (_Ktv2 *Ktv2Session) TotalBurned() (*big.Int, error) {
	return _Ktv2.Contract.TotalBurned(&_Ktv2.CallOpts)
}

// TotalBurned is a free data retrieval call binding the contract method 0xd89135cd.
//
// Solidity: function totalBurned() view returns(uint256)
func (_Ktv2 *Ktv2CallerSession) TotalBurned() (*big.Int, error) {
	return _Ktv2.Contract.TotalBurned(&_Ktv2.CallOpts)
}

// TotalGvn is a free data retrieval call binding the contract method 0x42935eb3.
//
// Solidity: function totalGvn() view returns(uint256)
func (_Ktv2 *Ktv2Caller) TotalGvn(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "totalGvn")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalGvn is a free data retrieval call binding the contract method 0x42935eb3.
//
// Solidity: function totalGvn() view returns(uint256)
func (_Ktv2 *Ktv2Session) TotalGvn() (*big.Int, error) {
	return _Ktv2.Contract.TotalGvn(&_Ktv2.CallOpts)
}

// TotalGvn is a free data retrieval call binding the contract method 0x42935eb3.
//
// Solidity: function totalGvn() view returns(uint256)
func (_Ktv2 *Ktv2CallerSession) TotalGvn() (*big.Int, error) {
	return _Ktv2.Contract.TotalGvn(&_Ktv2.CallOpts)
}

// TotalOC is a free data retrieval call binding the contract method 0x03c424c7.
//
// Solidity: function totalOC() view returns(uint16)
func (_Ktv2 *Ktv2Caller) TotalOC(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "totalOC")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// TotalOC is a free data retrieval call binding the contract method 0x03c424c7.
//
// Solidity: function totalOC() view returns(uint16)
func (_Ktv2 *Ktv2Session) TotalOC() (uint16, error) {
	return _Ktv2.Contract.TotalOC(&_Ktv2.CallOpts)
}

// TotalOC is a free data retrieval call binding the contract method 0x03c424c7.
//
// Solidity: function totalOC() view returns(uint16)
func (_Ktv2 *Ktv2CallerSession) TotalOC() (uint16, error) {
	return _Ktv2.Contract.TotalOC(&_Ktv2.CallOpts)
}

// TotalStk is a free data retrieval call binding the contract method 0x081a7ad3.
//
// Solidity: function totalStk() view returns(uint256)
func (_Ktv2 *Ktv2Caller) TotalStk(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "totalStk")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalStk is a free data retrieval call binding the contract method 0x081a7ad3.
//
// Solidity: function totalStk() view returns(uint256)
func (_Ktv2 *Ktv2Session) TotalStk() (*big.Int, error) {
	return _Ktv2.Contract.TotalStk(&_Ktv2.CallOpts)
}

// TotalStk is a free data retrieval call binding the contract method 0x081a7ad3.
//
// Solidity: function totalStk() view returns(uint256)
func (_Ktv2 *Ktv2CallerSession) TotalStk() (*big.Int, error) {
	return _Ktv2.Contract.TotalStk(&_Ktv2.CallOpts)
}

// Tp is a free data retrieval call binding the contract method 0x944de246.
//
// Solidity: function tp() view returns(address)
func (_Ktv2 *Ktv2Caller) Tp(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "tp")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Tp is a free data retrieval call binding the contract method 0x944de246.
//
// Solidity: function tp() view returns(address)
func (_Ktv2 *Ktv2Session) Tp() (common.Address, error) {
	return _Ktv2.Contract.Tp(&_Ktv2.CallOpts)
}

// Tp is a free data retrieval call binding the contract method 0x944de246.
//
// Solidity: function tp() view returns(address)
func (_Ktv2 *Ktv2CallerSession) Tp() (common.Address, error) {
	return _Ktv2.Contract.Tp(&_Ktv2.CallOpts)
}

// UserStks is a free data retrieval call binding the contract method 0x42d79e0d.
//
// Solidity: function userStks(address ) view returns(uint256)
func (_Ktv2 *Ktv2Caller) UserStks(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "userStks", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UserStks is a free data retrieval call binding the contract method 0x42d79e0d.
//
// Solidity: function userStks(address ) view returns(uint256)
func (_Ktv2 *Ktv2Session) UserStks(arg0 common.Address) (*big.Int, error) {
	return _Ktv2.Contract.UserStks(&_Ktv2.CallOpts, arg0)
}

// UserStks is a free data retrieval call binding the contract method 0x42d79e0d.
//
// Solidity: function userStks(address ) view returns(uint256)
func (_Ktv2 *Ktv2CallerSession) UserStks(arg0 common.Address) (*big.Int, error) {
	return _Ktv2.Contract.UserStks(&_Ktv2.CallOpts, arg0)
}

// V2 is a free data retrieval call binding the contract method 0xf3acae3a.
//
// Solidity: function v2() view returns(bool)
func (_Ktv2 *Ktv2Caller) V2(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Ktv2.contract.Call(opts, &out, "v2")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// V2 is a free data retrieval call binding the contract method 0xf3acae3a.
//
// Solidity: function v2() view returns(bool)
func (_Ktv2 *Ktv2Session) V2() (bool, error) {
	return _Ktv2.Contract.V2(&_Ktv2.CallOpts)
}

// V2 is a free data retrieval call binding the contract method 0xf3acae3a.
//
// Solidity: function v2() view returns(bool)
func (_Ktv2 *Ktv2CallerSession) V2() (bool, error) {
	return _Ktv2.Contract.V2(&_Ktv2.CallOpts)
}

// AddOCRwdr is a paid mutator transaction binding the contract method 0xd138321e.
//
// Solidity: function addOCRwdr(address addr) returns()
func (_Ktv2 *Ktv2Transactor) AddOCRwdr(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "addOCRwdr", addr)
}

// AddOCRwdr is a paid mutator transaction binding the contract method 0xd138321e.
//
// Solidity: function addOCRwdr(address addr) returns()
func (_Ktv2 *Ktv2Session) AddOCRwdr(addr common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.AddOCRwdr(&_Ktv2.TransactOpts, addr)
}

// AddOCRwdr is a paid mutator transaction binding the contract method 0xd138321e.
//
// Solidity: function addOCRwdr(address addr) returns()
func (_Ktv2 *Ktv2TransactorSession) AddOCRwdr(addr common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.AddOCRwdr(&_Ktv2.TransactOpts, addr)
}

// Allow is a paid mutator transaction binding the contract method 0xb1b3d3f6.
//
// Solidity: function allow() returns()
func (_Ktv2 *Ktv2Transactor) Allow(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "allow")
}

// Allow is a paid mutator transaction binding the contract method 0xb1b3d3f6.
//
// Solidity: function allow() returns()
func (_Ktv2 *Ktv2Session) Allow() (*types.Transaction, error) {
	return _Ktv2.Contract.Allow(&_Ktv2.TransactOpts)
}

// Allow is a paid mutator transaction binding the contract method 0xb1b3d3f6.
//
// Solidity: function allow() returns()
func (_Ktv2 *Ktv2TransactorSession) Allow() (*types.Transaction, error) {
	return _Ktv2.Contract.Allow(&_Ktv2.TransactOpts)
}

// Decline is a paid mutator transaction binding the contract method 0xab040107.
//
// Solidity: function decline() returns()
func (_Ktv2 *Ktv2Transactor) Decline(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "decline")
}

// Decline is a paid mutator transaction binding the contract method 0xab040107.
//
// Solidity: function decline() returns()
func (_Ktv2 *Ktv2Session) Decline() (*types.Transaction, error) {
	return _Ktv2.Contract.Decline(&_Ktv2.TransactOpts)
}

// Decline is a paid mutator transaction binding the contract method 0xab040107.
//
// Solidity: function decline() returns()
func (_Ktv2 *Ktv2TransactorSession) Decline() (*types.Transaction, error) {
	return _Ktv2.Contract.Decline(&_Ktv2.TransactOpts)
}

// Give is a paid mutator transaction binding the contract method 0x9e96a23a.
//
// Solidity: function give() payable returns()
func (_Ktv2 *Ktv2Transactor) Give(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "give")
}

// Give is a paid mutator transaction binding the contract method 0x9e96a23a.
//
// Solidity: function give() payable returns()
func (_Ktv2 *Ktv2Session) Give() (*types.Transaction, error) {
	return _Ktv2.Contract.Give(&_Ktv2.TransactOpts)
}

// Give is a paid mutator transaction binding the contract method 0x9e96a23a.
//
// Solidity: function give() payable returns()
func (_Ktv2 *Ktv2TransactorSession) Give() (*types.Transaction, error) {
	return _Ktv2.Contract.Give(&_Ktv2.TransactOpts)
}

// RemoveOCRwdr is a paid mutator transaction binding the contract method 0xadec2f77.
//
// Solidity: function removeOCRwdr(address addr) returns()
func (_Ktv2 *Ktv2Transactor) RemoveOCRwdr(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "removeOCRwdr", addr)
}

// RemoveOCRwdr is a paid mutator transaction binding the contract method 0xadec2f77.
//
// Solidity: function removeOCRwdr(address addr) returns()
func (_Ktv2 *Ktv2Session) RemoveOCRwdr(addr common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.RemoveOCRwdr(&_Ktv2.TransactOpts, addr)
}

// RemoveOCRwdr is a paid mutator transaction binding the contract method 0xadec2f77.
//
// Solidity: function removeOCRwdr(address addr) returns()
func (_Ktv2 *Ktv2TransactorSession) RemoveOCRwdr(addr common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.RemoveOCRwdr(&_Ktv2.TransactOpts, addr)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ktv2 *Ktv2Transactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ktv2 *Ktv2Session) RenounceOwnership() (*types.Transaction, error) {
	return _Ktv2.Contract.RenounceOwnership(&_Ktv2.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ktv2 *Ktv2TransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Ktv2.Contract.RenounceOwnership(&_Ktv2.TransactOpts)
}

// ResetVote is a paid mutator transaction binding the contract method 0x2d4812ed.
//
// Solidity: function resetVote(address _to) returns()
func (_Ktv2 *Ktv2Transactor) ResetVote(opts *bind.TransactOpts, _to common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "resetVote", _to)
}

// ResetVote is a paid mutator transaction binding the contract method 0x2d4812ed.
//
// Solidity: function resetVote(address _to) returns()
func (_Ktv2 *Ktv2Session) ResetVote(_to common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.ResetVote(&_Ktv2.TransactOpts, _to)
}

// ResetVote is a paid mutator transaction binding the contract method 0x2d4812ed.
//
// Solidity: function resetVote(address _to) returns()
func (_Ktv2 *Ktv2TransactorSession) ResetVote(_to common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.ResetVote(&_Ktv2.TransactOpts, _to)
}

// ResetVoteToAdd is a paid mutator transaction binding the contract method 0xb53da9d9.
//
// Solidity: function resetVoteToAdd(address newOC) returns()
func (_Ktv2 *Ktv2Transactor) ResetVoteToAdd(opts *bind.TransactOpts, newOC common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "resetVoteToAdd", newOC)
}

// ResetVoteToAdd is a paid mutator transaction binding the contract method 0xb53da9d9.
//
// Solidity: function resetVoteToAdd(address newOC) returns()
func (_Ktv2 *Ktv2Session) ResetVoteToAdd(newOC common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.ResetVoteToAdd(&_Ktv2.TransactOpts, newOC)
}

// ResetVoteToAdd is a paid mutator transaction binding the contract method 0xb53da9d9.
//
// Solidity: function resetVoteToAdd(address newOC) returns()
func (_Ktv2 *Ktv2TransactorSession) ResetVoteToAdd(newOC common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.ResetVoteToAdd(&_Ktv2.TransactOpts, newOC)
}

// ResetVoteToRemove is a paid mutator transaction binding the contract method 0xdfa00b1c.
//
// Solidity: function resetVoteToRemove(address existingOC) returns()
func (_Ktv2 *Ktv2Transactor) ResetVoteToRemove(opts *bind.TransactOpts, existingOC common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "resetVoteToRemove", existingOC)
}

// ResetVoteToRemove is a paid mutator transaction binding the contract method 0xdfa00b1c.
//
// Solidity: function resetVoteToRemove(address existingOC) returns()
func (_Ktv2 *Ktv2Session) ResetVoteToRemove(existingOC common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.ResetVoteToRemove(&_Ktv2.TransactOpts, existingOC)
}

// ResetVoteToRemove is a paid mutator transaction binding the contract method 0xdfa00b1c.
//
// Solidity: function resetVoteToRemove(address existingOC) returns()
func (_Ktv2 *Ktv2TransactorSession) ResetVoteToRemove(existingOC common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.ResetVoteToRemove(&_Ktv2.TransactOpts, existingOC)
}

// Rwd is a paid mutator transaction binding the contract method 0x30606eaf.
//
// Solidity: function rwd(address _to, uint256 _amt) returns()
func (_Ktv2 *Ktv2Transactor) Rwd(opts *bind.TransactOpts, _to common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "rwd", _to, _amt)
}

// Rwd is a paid mutator transaction binding the contract method 0x30606eaf.
//
// Solidity: function rwd(address _to, uint256 _amt) returns()
func (_Ktv2 *Ktv2Session) Rwd(_to common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.Contract.Rwd(&_Ktv2.TransactOpts, _to, _amt)
}

// Rwd is a paid mutator transaction binding the contract method 0x30606eaf.
//
// Solidity: function rwd(address _to, uint256 _amt) returns()
func (_Ktv2 *Ktv2TransactorSession) Rwd(_to common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.Contract.Rwd(&_Ktv2.TransactOpts, _to, _amt)
}

// SetBurnFactor is a paid mutator transaction binding the contract method 0x3fdf4f62.
//
// Solidity: function setBurnFactor(uint16 amt) returns()
func (_Ktv2 *Ktv2Transactor) SetBurnFactor(opts *bind.TransactOpts, amt uint16) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setBurnFactor", amt)
}

// SetBurnFactor is a paid mutator transaction binding the contract method 0x3fdf4f62.
//
// Solidity: function setBurnFactor(uint16 amt) returns()
func (_Ktv2 *Ktv2Session) SetBurnFactor(amt uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetBurnFactor(&_Ktv2.TransactOpts, amt)
}

// SetBurnFactor is a paid mutator transaction binding the contract method 0x3fdf4f62.
//
// Solidity: function setBurnFactor(uint16 amt) returns()
func (_Ktv2 *Ktv2TransactorSession) SetBurnFactor(amt uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetBurnFactor(&_Ktv2.TransactOpts, amt)
}

// SetConsensusReq is a paid mutator transaction binding the contract method 0xab3ad1d4.
//
// Solidity: function setConsensusReq(uint16 req) returns()
func (_Ktv2 *Ktv2Transactor) SetConsensusReq(opts *bind.TransactOpts, req uint16) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setConsensusReq", req)
}

// SetConsensusReq is a paid mutator transaction binding the contract method 0xab3ad1d4.
//
// Solidity: function setConsensusReq(uint16 req) returns()
func (_Ktv2 *Ktv2Session) SetConsensusReq(req uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetConsensusReq(&_Ktv2.TransactOpts, req)
}

// SetConsensusReq is a paid mutator transaction binding the contract method 0xab3ad1d4.
//
// Solidity: function setConsensusReq(uint16 req) returns()
func (_Ktv2 *Ktv2TransactorSession) SetConsensusReq(req uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetConsensusReq(&_Ktv2.TransactOpts, req)
}

// SetDest is a paid mutator transaction binding the contract method 0x1b46ba84.
//
// Solidity: function setDest(address addr) returns()
func (_Ktv2 *Ktv2Transactor) SetDest(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setDest", addr)
}

// SetDest is a paid mutator transaction binding the contract method 0x1b46ba84.
//
// Solidity: function setDest(address addr) returns()
func (_Ktv2 *Ktv2Session) SetDest(addr common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.SetDest(&_Ktv2.TransactOpts, addr)
}

// SetDest is a paid mutator transaction binding the contract method 0x1b46ba84.
//
// Solidity: function setDest(address addr) returns()
func (_Ktv2 *Ktv2TransactorSession) SetDest(addr common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.SetDest(&_Ktv2.TransactOpts, addr)
}

// SetDonationPrc is a paid mutator transaction binding the contract method 0xfd52b702.
//
// Solidity: function setDonationPrc(uint16 amt) returns()
func (_Ktv2 *Ktv2Transactor) SetDonationPrc(opts *bind.TransactOpts, amt uint16) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setDonationPrc", amt)
}

// SetDonationPrc is a paid mutator transaction binding the contract method 0xfd52b702.
//
// Solidity: function setDonationPrc(uint16 amt) returns()
func (_Ktv2 *Ktv2Session) SetDonationPrc(amt uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetDonationPrc(&_Ktv2.TransactOpts, amt)
}

// SetDonationPrc is a paid mutator transaction binding the contract method 0xfd52b702.
//
// Solidity: function setDonationPrc(uint16 amt) returns()
func (_Ktv2 *Ktv2TransactorSession) SetDonationPrc(amt uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetDonationPrc(&_Ktv2.TransactOpts, amt)
}

// SetEpochInterval is a paid mutator transaction binding the contract method 0x3407e1c3.
//
// Solidity: function setEpochInterval(uint16 interval) returns()
func (_Ktv2 *Ktv2Transactor) SetEpochInterval(opts *bind.TransactOpts, interval uint16) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setEpochInterval", interval)
}

// SetEpochInterval is a paid mutator transaction binding the contract method 0x3407e1c3.
//
// Solidity: function setEpochInterval(uint16 interval) returns()
func (_Ktv2 *Ktv2Session) SetEpochInterval(interval uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetEpochInterval(&_Ktv2.TransactOpts, interval)
}

// SetEpochInterval is a paid mutator transaction binding the contract method 0x3407e1c3.
//
// Solidity: function setEpochInterval(uint16 interval) returns()
func (_Ktv2 *Ktv2TransactorSession) SetEpochInterval(interval uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetEpochInterval(&_Ktv2.TransactOpts, interval)
}

// SetMaxBurnPrc is a paid mutator transaction binding the contract method 0xe89d048e.
//
// Solidity: function setMaxBurnPrc(uint16 amt) returns()
func (_Ktv2 *Ktv2Transactor) SetMaxBurnPrc(opts *bind.TransactOpts, amt uint16) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setMaxBurnPrc", amt)
}

// SetMaxBurnPrc is a paid mutator transaction binding the contract method 0xe89d048e.
//
// Solidity: function setMaxBurnPrc(uint16 amt) returns()
func (_Ktv2 *Ktv2Session) SetMaxBurnPrc(amt uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetMaxBurnPrc(&_Ktv2.TransactOpts, amt)
}

// SetMaxBurnPrc is a paid mutator transaction binding the contract method 0xe89d048e.
//
// Solidity: function setMaxBurnPrc(uint16 amt) returns()
func (_Ktv2 *Ktv2TransactorSession) SetMaxBurnPrc(amt uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetMaxBurnPrc(&_Ktv2.TransactOpts, amt)
}

// SetOCFee is a paid mutator transaction binding the contract method 0xc377e65d.
//
// Solidity: function setOCFee(uint16 fee) returns()
func (_Ktv2 *Ktv2Transactor) SetOCFee(opts *bind.TransactOpts, fee uint16) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setOCFee", fee)
}

// SetOCFee is a paid mutator transaction binding the contract method 0xc377e65d.
//
// Solidity: function setOCFee(uint16 fee) returns()
func (_Ktv2 *Ktv2Session) SetOCFee(fee uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetOCFee(&_Ktv2.TransactOpts, fee)
}

// SetOCFee is a paid mutator transaction binding the contract method 0xc377e65d.
//
// Solidity: function setOCFee(uint16 fee) returns()
func (_Ktv2 *Ktv2TransactorSession) SetOCFee(fee uint16) (*types.Transaction, error) {
	return _Ktv2.Contract.SetOCFee(&_Ktv2.TransactOpts, fee)
}

// SetPool is a paid mutator transaction binding the contract method 0x4437152a.
//
// Solidity: function setPool(address _pool) returns()
func (_Ktv2 *Ktv2Transactor) SetPool(opts *bind.TransactOpts, _pool common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setPool", _pool)
}

// SetPool is a paid mutator transaction binding the contract method 0x4437152a.
//
// Solidity: function setPool(address _pool) returns()
func (_Ktv2 *Ktv2Session) SetPool(_pool common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.SetPool(&_Ktv2.TransactOpts, _pool)
}

// SetPool is a paid mutator transaction binding the contract method 0x4437152a.
//
// Solidity: function setPool(address _pool) returns()
func (_Ktv2 *Ktv2TransactorSession) SetPool(_pool common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.SetPool(&_Ktv2.TransactOpts, _pool)
}

// SetV2 is a paid mutator transaction binding the contract method 0x30e3cb68.
//
// Solidity: function setV2(bool _v2) returns()
func (_Ktv2 *Ktv2Transactor) SetV2(opts *bind.TransactOpts, _v2 bool) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "setV2", _v2)
}

// SetV2 is a paid mutator transaction binding the contract method 0x30e3cb68.
//
// Solidity: function setV2(bool _v2) returns()
func (_Ktv2 *Ktv2Session) SetV2(_v2 bool) (*types.Transaction, error) {
	return _Ktv2.Contract.SetV2(&_Ktv2.TransactOpts, _v2)
}

// SetV2 is a paid mutator transaction binding the contract method 0x30e3cb68.
//
// Solidity: function setV2(bool _v2) returns()
func (_Ktv2 *Ktv2TransactorSession) SetV2(_v2 bool) (*types.Transaction, error) {
	return _Ktv2.Contract.SetV2(&_Ktv2.TransactOpts, _v2)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 amt) returns()
func (_Ktv2 *Ktv2Transactor) Stake(opts *bind.TransactOpts, amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "stake", amt)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 amt) returns()
func (_Ktv2 *Ktv2Session) Stake(amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.Contract.Stake(&_Ktv2.TransactOpts, amt)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 amt) returns()
func (_Ktv2 *Ktv2TransactorSession) Stake(amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.Contract.Stake(&_Ktv2.TransactOpts, amt)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ktv2 *Ktv2Transactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ktv2 *Ktv2Session) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.TransferOwnership(&_Ktv2.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ktv2 *Ktv2TransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.TransferOwnership(&_Ktv2.TransactOpts, newOwner)
}

// Vote is a paid mutator transaction binding the contract method 0x1bf533a4.
//
// Solidity: function vote(address _to, string data) returns()
func (_Ktv2 *Ktv2Transactor) Vote(opts *bind.TransactOpts, _to common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "vote", _to, data)
}

// Vote is a paid mutator transaction binding the contract method 0x1bf533a4.
//
// Solidity: function vote(address _to, string data) returns()
func (_Ktv2 *Ktv2Session) Vote(_to common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.Contract.Vote(&_Ktv2.TransactOpts, _to, data)
}

// Vote is a paid mutator transaction binding the contract method 0x1bf533a4.
//
// Solidity: function vote(address _to, string data) returns()
func (_Ktv2 *Ktv2TransactorSession) Vote(_to common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.Contract.Vote(&_Ktv2.TransactOpts, _to, data)
}

// VoteToAdd is a paid mutator transaction binding the contract method 0x7b0443fa.
//
// Solidity: function voteToAdd(address newOC, string data) returns()
func (_Ktv2 *Ktv2Transactor) VoteToAdd(opts *bind.TransactOpts, newOC common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "voteToAdd", newOC, data)
}

// VoteToAdd is a paid mutator transaction binding the contract method 0x7b0443fa.
//
// Solidity: function voteToAdd(address newOC, string data) returns()
func (_Ktv2 *Ktv2Session) VoteToAdd(newOC common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.Contract.VoteToAdd(&_Ktv2.TransactOpts, newOC, data)
}

// VoteToAdd is a paid mutator transaction binding the contract method 0x7b0443fa.
//
// Solidity: function voteToAdd(address newOC, string data) returns()
func (_Ktv2 *Ktv2TransactorSession) VoteToAdd(newOC common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.Contract.VoteToAdd(&_Ktv2.TransactOpts, newOC, data)
}

// VoteToRemove is a paid mutator transaction binding the contract method 0xca1d9cc2.
//
// Solidity: function voteToRemove(address existingOC, string data) returns()
func (_Ktv2 *Ktv2Transactor) VoteToRemove(opts *bind.TransactOpts, existingOC common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "voteToRemove", existingOC, data)
}

// VoteToRemove is a paid mutator transaction binding the contract method 0xca1d9cc2.
//
// Solidity: function voteToRemove(address existingOC, string data) returns()
func (_Ktv2 *Ktv2Session) VoteToRemove(existingOC common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.Contract.VoteToRemove(&_Ktv2.TransactOpts, existingOC, data)
}

// VoteToRemove is a paid mutator transaction binding the contract method 0xca1d9cc2.
//
// Solidity: function voteToRemove(address existingOC, string data) returns()
func (_Ktv2 *Ktv2TransactorSession) VoteToRemove(existingOC common.Address, data string) (*types.Transaction, error) {
	return _Ktv2.Contract.VoteToRemove(&_Ktv2.TransactOpts, existingOC, data)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amt) returns()
func (_Ktv2 *Ktv2Transactor) Withdraw(opts *bind.TransactOpts, amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "withdraw", amt)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amt) returns()
func (_Ktv2 *Ktv2Session) Withdraw(amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.Contract.Withdraw(&_Ktv2.TransactOpts, amt)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amt) returns()
func (_Ktv2 *Ktv2TransactorSession) Withdraw(amt *big.Int) (*types.Transaction, error) {
	return _Ktv2.Contract.Withdraw(&_Ktv2.TransactOpts, amt)
}

// WithdrawOCFee is a paid mutator transaction binding the contract method 0xf14f4529.
//
// Solidity: function withdrawOCFee(uint32[] blocks) returns()
func (_Ktv2 *Ktv2Transactor) WithdrawOCFee(opts *bind.TransactOpts, blocks []uint32) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "withdrawOCFee", blocks)
}

// WithdrawOCFee is a paid mutator transaction binding the contract method 0xf14f4529.
//
// Solidity: function withdrawOCFee(uint32[] blocks) returns()
func (_Ktv2 *Ktv2Session) WithdrawOCFee(blocks []uint32) (*types.Transaction, error) {
	return _Ktv2.Contract.WithdrawOCFee(&_Ktv2.TransactOpts, blocks)
}

// WithdrawOCFee is a paid mutator transaction binding the contract method 0xf14f4529.
//
// Solidity: function withdrawOCFee(uint32[] blocks) returns()
func (_Ktv2 *Ktv2TransactorSession) WithdrawOCFee(blocks []uint32) (*types.Transaction, error) {
	return _Ktv2.Contract.WithdrawOCFee(&_Ktv2.TransactOpts, blocks)
}

// WithdrawTkn is a paid mutator transaction binding the contract method 0xa42bebbe.
//
// Solidity: function withdrawTkn(address _to, address addr) returns()
func (_Ktv2 *Ktv2Transactor) WithdrawTkn(opts *bind.TransactOpts, _to common.Address, addr common.Address) (*types.Transaction, error) {
	return _Ktv2.contract.Transact(opts, "withdrawTkn", _to, addr)
}

// WithdrawTkn is a paid mutator transaction binding the contract method 0xa42bebbe.
//
// Solidity: function withdrawTkn(address _to, address addr) returns()
func (_Ktv2 *Ktv2Session) WithdrawTkn(_to common.Address, addr common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.WithdrawTkn(&_Ktv2.TransactOpts, _to, addr)
}

// WithdrawTkn is a paid mutator transaction binding the contract method 0xa42bebbe.
//
// Solidity: function withdrawTkn(address _to, address addr) returns()
func (_Ktv2 *Ktv2TransactorSession) WithdrawTkn(_to common.Address, addr common.Address) (*types.Transaction, error) {
	return _Ktv2.Contract.WithdrawTkn(&_Ktv2.TransactOpts, _to, addr)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Ktv2 *Ktv2Transactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Ktv2 *Ktv2Session) Receive() (*types.Transaction, error) {
	return _Ktv2.Contract.Receive(&_Ktv2.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Ktv2 *Ktv2TransactorSession) Receive() (*types.Transaction, error) {
	return _Ktv2.Contract.Receive(&_Ktv2.TransactOpts)
}

// Ktv2GaveIterator is returned from FilterGave and is used to iterate over the raw logs and unpacked data for Gave events raised by the Ktv2 contract.
type Ktv2GaveIterator struct {
	Event *Ktv2Gave // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2GaveIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2Gave)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2Gave)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2GaveIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2GaveIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2Gave represents a Gave event raised by the Ktv2 contract.
type Ktv2Gave struct {
	Arg0 common.Address
	Arg1 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterGave is a free log retrieval operation binding the contract event 0xe3ba4a7522b6c3133015e07f410a43323021989ceedd5de9d51c5ec155288974.
//
// Solidity: event Gave(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) FilterGave(opts *bind.FilterOpts) (*Ktv2GaveIterator, error) {

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "Gave")
	if err != nil {
		return nil, err
	}
	return &Ktv2GaveIterator{contract: _Ktv2.contract, event: "Gave", logs: logs, sub: sub}, nil
}

// WatchGave is a free log subscription operation binding the contract event 0xe3ba4a7522b6c3133015e07f410a43323021989ceedd5de9d51c5ec155288974.
//
// Solidity: event Gave(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) WatchGave(opts *bind.WatchOpts, sink chan<- *Ktv2Gave) (event.Subscription, error) {

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "Gave")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2Gave)
				if err := _Ktv2.contract.UnpackLog(event, "Gave", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseGave is a log parse operation binding the contract event 0xe3ba4a7522b6c3133015e07f410a43323021989ceedd5de9d51c5ec155288974.
//
// Solidity: event Gave(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) ParseGave(log types.Log) (*Ktv2Gave, error) {
	event := new(Ktv2Gave)
	if err := _Ktv2.contract.UnpackLog(event, "Gave", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2NodeAddedIterator is returned from FilterNodeAdded and is used to iterate over the raw logs and unpacked data for NodeAdded events raised by the Ktv2 contract.
type Ktv2NodeAddedIterator struct {
	Event *Ktv2NodeAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2NodeAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2NodeAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2NodeAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2NodeAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2NodeAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2NodeAdded represents a NodeAdded event raised by the Ktv2 contract.
type Ktv2NodeAdded struct {
	NewOC common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNodeAdded is a free log retrieval operation binding the contract event 0xb25d03aaf308d7291709be1ea28b800463cf3a9a4c4a5555d7333a964c1dfebd.
//
// Solidity: event NodeAdded(address indexed newOC)
func (_Ktv2 *Ktv2Filterer) FilterNodeAdded(opts *bind.FilterOpts, newOC []common.Address) (*Ktv2NodeAddedIterator, error) {

	var newOCRule []interface{}
	for _, newOCItem := range newOC {
		newOCRule = append(newOCRule, newOCItem)
	}

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "NodeAdded", newOCRule)
	if err != nil {
		return nil, err
	}
	return &Ktv2NodeAddedIterator{contract: _Ktv2.contract, event: "NodeAdded", logs: logs, sub: sub}, nil
}

// WatchNodeAdded is a free log subscription operation binding the contract event 0xb25d03aaf308d7291709be1ea28b800463cf3a9a4c4a5555d7333a964c1dfebd.
//
// Solidity: event NodeAdded(address indexed newOC)
func (_Ktv2 *Ktv2Filterer) WatchNodeAdded(opts *bind.WatchOpts, sink chan<- *Ktv2NodeAdded, newOC []common.Address) (event.Subscription, error) {

	var newOCRule []interface{}
	for _, newOCItem := range newOC {
		newOCRule = append(newOCRule, newOCItem)
	}

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "NodeAdded", newOCRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2NodeAdded)
				if err := _Ktv2.contract.UnpackLog(event, "NodeAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNodeAdded is a log parse operation binding the contract event 0xb25d03aaf308d7291709be1ea28b800463cf3a9a4c4a5555d7333a964c1dfebd.
//
// Solidity: event NodeAdded(address indexed newOC)
func (_Ktv2 *Ktv2Filterer) ParseNodeAdded(log types.Log) (*Ktv2NodeAdded, error) {
	event := new(Ktv2NodeAdded)
	if err := _Ktv2.contract.UnpackLog(event, "NodeAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2NodeRemovedIterator is returned from FilterNodeRemoved and is used to iterate over the raw logs and unpacked data for NodeRemoved events raised by the Ktv2 contract.
type Ktv2NodeRemovedIterator struct {
	Event *Ktv2NodeRemoved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2NodeRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2NodeRemoved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2NodeRemoved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2NodeRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2NodeRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2NodeRemoved represents a NodeRemoved event raised by the Ktv2 contract.
type Ktv2NodeRemoved struct {
	OldOC common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNodeRemoved is a free log retrieval operation binding the contract event 0xcfc24166db4bb677e857cacabd1541fb2b30645021b27c5130419589b84db52b.
//
// Solidity: event NodeRemoved(address indexed oldOC)
func (_Ktv2 *Ktv2Filterer) FilterNodeRemoved(opts *bind.FilterOpts, oldOC []common.Address) (*Ktv2NodeRemovedIterator, error) {

	var oldOCRule []interface{}
	for _, oldOCItem := range oldOC {
		oldOCRule = append(oldOCRule, oldOCItem)
	}

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "NodeRemoved", oldOCRule)
	if err != nil {
		return nil, err
	}
	return &Ktv2NodeRemovedIterator{contract: _Ktv2.contract, event: "NodeRemoved", logs: logs, sub: sub}, nil
}

// WatchNodeRemoved is a free log subscription operation binding the contract event 0xcfc24166db4bb677e857cacabd1541fb2b30645021b27c5130419589b84db52b.
//
// Solidity: event NodeRemoved(address indexed oldOC)
func (_Ktv2 *Ktv2Filterer) WatchNodeRemoved(opts *bind.WatchOpts, sink chan<- *Ktv2NodeRemoved, oldOC []common.Address) (event.Subscription, error) {

	var oldOCRule []interface{}
	for _, oldOCItem := range oldOC {
		oldOCRule = append(oldOCRule, oldOCItem)
	}

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "NodeRemoved", oldOCRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2NodeRemoved)
				if err := _Ktv2.contract.UnpackLog(event, "NodeRemoved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNodeRemoved is a log parse operation binding the contract event 0xcfc24166db4bb677e857cacabd1541fb2b30645021b27c5130419589b84db52b.
//
// Solidity: event NodeRemoved(address indexed oldOC)
func (_Ktv2 *Ktv2Filterer) ParseNodeRemoved(log types.Log) (*Ktv2NodeRemoved, error) {
	event := new(Ktv2NodeRemoved)
	if err := _Ktv2.contract.UnpackLog(event, "NodeRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2OwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Ktv2 contract.
type Ktv2OwnershipTransferredIterator struct {
	Event *Ktv2OwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2OwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2OwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2OwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2OwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2OwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2OwnershipTransferred represents a OwnershipTransferred event raised by the Ktv2 contract.
type Ktv2OwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Ktv2 *Ktv2Filterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*Ktv2OwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &Ktv2OwnershipTransferredIterator{contract: _Ktv2.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Ktv2 *Ktv2Filterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *Ktv2OwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2OwnershipTransferred)
				if err := _Ktv2.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Ktv2 *Ktv2Filterer) ParseOwnershipTransferred(log types.Log) (*Ktv2OwnershipTransferred, error) {
	event := new(Ktv2OwnershipTransferred)
	if err := _Ktv2.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2RwdIterator is returned from FilterRwd and is used to iterate over the raw logs and unpacked data for Rwd events raised by the Ktv2 contract.
type Ktv2RwdIterator struct {
	Event *Ktv2Rwd // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2RwdIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2Rwd)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2Rwd)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2RwdIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2RwdIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2Rwd represents a Rwd event raised by the Ktv2 contract.
type Ktv2Rwd struct {
	Arg0 common.Address
	Arg1 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRwd is a free log retrieval operation binding the contract event 0x5df9aa58816b17fb728eddad0162659653310b4f9cd8796a7c647612b3d520ba.
//
// Solidity: event Rwd(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) FilterRwd(opts *bind.FilterOpts) (*Ktv2RwdIterator, error) {

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "Rwd")
	if err != nil {
		return nil, err
	}
	return &Ktv2RwdIterator{contract: _Ktv2.contract, event: "Rwd", logs: logs, sub: sub}, nil
}

// WatchRwd is a free log subscription operation binding the contract event 0x5df9aa58816b17fb728eddad0162659653310b4f9cd8796a7c647612b3d520ba.
//
// Solidity: event Rwd(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) WatchRwd(opts *bind.WatchOpts, sink chan<- *Ktv2Rwd) (event.Subscription, error) {

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "Rwd")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2Rwd)
				if err := _Ktv2.contract.UnpackLog(event, "Rwd", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRwd is a log parse operation binding the contract event 0x5df9aa58816b17fb728eddad0162659653310b4f9cd8796a7c647612b3d520ba.
//
// Solidity: event Rwd(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) ParseRwd(log types.Log) (*Ktv2Rwd, error) {
	event := new(Ktv2Rwd)
	if err := _Ktv2.contract.UnpackLog(event, "Rwd", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2StakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the Ktv2 contract.
type Ktv2StakedIterator struct {
	Event *Ktv2Staked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2StakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2Staked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2Staked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2StakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2StakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2Staked represents a Staked event raised by the Ktv2 contract.
type Ktv2Staked struct {
	Arg0 common.Address
	Arg1 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) FilterStaked(opts *bind.FilterOpts) (*Ktv2StakedIterator, error) {

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "Staked")
	if err != nil {
		return nil, err
	}
	return &Ktv2StakedIterator{contract: _Ktv2.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *Ktv2Staked) (event.Subscription, error) {

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "Staked")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2Staked)
				if err := _Ktv2.contract.UnpackLog(event, "Staked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStaked is a log parse operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) ParseStaked(log types.Log) (*Ktv2Staked, error) {
	event := new(Ktv2Staked)
	if err := _Ktv2.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2VotedIterator is returned from FilterVoted and is used to iterate over the raw logs and unpacked data for Voted events raised by the Ktv2 contract.
type Ktv2VotedIterator struct {
	Event *Ktv2Voted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2VotedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2Voted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2Voted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2VotedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2VotedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2Voted represents a Voted event raised by the Ktv2 contract.
type Ktv2Voted struct {
	Arg0 *big.Int
	Arg1 common.Address
	Arg2 string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterVoted is a free log retrieval operation binding the contract event 0x909b43dcc56d91024768bbc5c8d56441234580b2fd6176844960cbc7218cc0b5.
//
// Solidity: event Voted(uint256 arg0, address arg1, string arg2)
func (_Ktv2 *Ktv2Filterer) FilterVoted(opts *bind.FilterOpts) (*Ktv2VotedIterator, error) {

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "Voted")
	if err != nil {
		return nil, err
	}
	return &Ktv2VotedIterator{contract: _Ktv2.contract, event: "Voted", logs: logs, sub: sub}, nil
}

// WatchVoted is a free log subscription operation binding the contract event 0x909b43dcc56d91024768bbc5c8d56441234580b2fd6176844960cbc7218cc0b5.
//
// Solidity: event Voted(uint256 arg0, address arg1, string arg2)
func (_Ktv2 *Ktv2Filterer) WatchVoted(opts *bind.WatchOpts, sink chan<- *Ktv2Voted) (event.Subscription, error) {

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "Voted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2Voted)
				if err := _Ktv2.contract.UnpackLog(event, "Voted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVoted is a log parse operation binding the contract event 0x909b43dcc56d91024768bbc5c8d56441234580b2fd6176844960cbc7218cc0b5.
//
// Solidity: event Voted(uint256 arg0, address arg1, string arg2)
func (_Ktv2 *Ktv2Filterer) ParseVoted(log types.Log) (*Ktv2Voted, error) {
	event := new(Ktv2Voted)
	if err := _Ktv2.contract.UnpackLog(event, "Voted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2VotedToAddIterator is returned from FilterVotedToAdd and is used to iterate over the raw logs and unpacked data for VotedToAdd events raised by the Ktv2 contract.
type Ktv2VotedToAddIterator struct {
	Event *Ktv2VotedToAdd // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2VotedToAddIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2VotedToAdd)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2VotedToAdd)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2VotedToAddIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2VotedToAddIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2VotedToAdd represents a VotedToAdd event raised by the Ktv2 contract.
type Ktv2VotedToAdd struct {
	Voter common.Address
	NewOC common.Address
	Data  string
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterVotedToAdd is a free log retrieval operation binding the contract event 0x6f16cedcfba9076088f50cb98393296e065155864bc7f3c923530550e35fd455.
//
// Solidity: event VotedToAdd(address indexed voter, address indexed newOC, string data)
func (_Ktv2 *Ktv2Filterer) FilterVotedToAdd(opts *bind.FilterOpts, voter []common.Address, newOC []common.Address) (*Ktv2VotedToAddIterator, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}
	var newOCRule []interface{}
	for _, newOCItem := range newOC {
		newOCRule = append(newOCRule, newOCItem)
	}

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "VotedToAdd", voterRule, newOCRule)
	if err != nil {
		return nil, err
	}
	return &Ktv2VotedToAddIterator{contract: _Ktv2.contract, event: "VotedToAdd", logs: logs, sub: sub}, nil
}

// WatchVotedToAdd is a free log subscription operation binding the contract event 0x6f16cedcfba9076088f50cb98393296e065155864bc7f3c923530550e35fd455.
//
// Solidity: event VotedToAdd(address indexed voter, address indexed newOC, string data)
func (_Ktv2 *Ktv2Filterer) WatchVotedToAdd(opts *bind.WatchOpts, sink chan<- *Ktv2VotedToAdd, voter []common.Address, newOC []common.Address) (event.Subscription, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}
	var newOCRule []interface{}
	for _, newOCItem := range newOC {
		newOCRule = append(newOCRule, newOCItem)
	}

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "VotedToAdd", voterRule, newOCRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2VotedToAdd)
				if err := _Ktv2.contract.UnpackLog(event, "VotedToAdd", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVotedToAdd is a log parse operation binding the contract event 0x6f16cedcfba9076088f50cb98393296e065155864bc7f3c923530550e35fd455.
//
// Solidity: event VotedToAdd(address indexed voter, address indexed newOC, string data)
func (_Ktv2 *Ktv2Filterer) ParseVotedToAdd(log types.Log) (*Ktv2VotedToAdd, error) {
	event := new(Ktv2VotedToAdd)
	if err := _Ktv2.contract.UnpackLog(event, "VotedToAdd", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2VotedToRemoveIterator is returned from FilterVotedToRemove and is used to iterate over the raw logs and unpacked data for VotedToRemove events raised by the Ktv2 contract.
type Ktv2VotedToRemoveIterator struct {
	Event *Ktv2VotedToRemove // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2VotedToRemoveIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2VotedToRemove)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2VotedToRemove)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2VotedToRemoveIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2VotedToRemoveIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2VotedToRemove represents a VotedToRemove event raised by the Ktv2 contract.
type Ktv2VotedToRemove struct {
	Voter      common.Address
	ExistingOC common.Address
	Data       string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterVotedToRemove is a free log retrieval operation binding the contract event 0x326236c4b24ce4385364a6e9ea51a988889c12b6ac3cea126fd823a9c822994f.
//
// Solidity: event VotedToRemove(address indexed voter, address indexed existingOC, string data)
func (_Ktv2 *Ktv2Filterer) FilterVotedToRemove(opts *bind.FilterOpts, voter []common.Address, existingOC []common.Address) (*Ktv2VotedToRemoveIterator, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}
	var existingOCRule []interface{}
	for _, existingOCItem := range existingOC {
		existingOCRule = append(existingOCRule, existingOCItem)
	}

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "VotedToRemove", voterRule, existingOCRule)
	if err != nil {
		return nil, err
	}
	return &Ktv2VotedToRemoveIterator{contract: _Ktv2.contract, event: "VotedToRemove", logs: logs, sub: sub}, nil
}

// WatchVotedToRemove is a free log subscription operation binding the contract event 0x326236c4b24ce4385364a6e9ea51a988889c12b6ac3cea126fd823a9c822994f.
//
// Solidity: event VotedToRemove(address indexed voter, address indexed existingOC, string data)
func (_Ktv2 *Ktv2Filterer) WatchVotedToRemove(opts *bind.WatchOpts, sink chan<- *Ktv2VotedToRemove, voter []common.Address, existingOC []common.Address) (event.Subscription, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}
	var existingOCRule []interface{}
	for _, existingOCItem := range existingOC {
		existingOCRule = append(existingOCRule, existingOCItem)
	}

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "VotedToRemove", voterRule, existingOCRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2VotedToRemove)
				if err := _Ktv2.contract.UnpackLog(event, "VotedToRemove", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVotedToRemove is a log parse operation binding the contract event 0x326236c4b24ce4385364a6e9ea51a988889c12b6ac3cea126fd823a9c822994f.
//
// Solidity: event VotedToRemove(address indexed voter, address indexed existingOC, string data)
func (_Ktv2 *Ktv2Filterer) ParseVotedToRemove(log types.Log) (*Ktv2VotedToRemove, error) {
	event := new(Ktv2VotedToRemove)
	if err := _Ktv2.contract.UnpackLog(event, "VotedToRemove", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Ktv2WithdrewIterator is returned from FilterWithdrew and is used to iterate over the raw logs and unpacked data for Withdrew events raised by the Ktv2 contract.
type Ktv2WithdrewIterator struct {
	Event *Ktv2Withdrew // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Ktv2WithdrewIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2Withdrew)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Ktv2Withdrew)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Ktv2WithdrewIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2WithdrewIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2Withdrew represents a Withdrew event raised by the Ktv2 contract.
type Ktv2Withdrew struct {
	Arg0 common.Address
	Arg1 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterWithdrew is a free log retrieval operation binding the contract event 0xb244b9a17ad633c6e83b7983ee04320484956a68ddbe96a0b70dfca1cf19d723.
//
// Solidity: event Withdrew(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) FilterWithdrew(opts *bind.FilterOpts) (*Ktv2WithdrewIterator, error) {

	logs, sub, err := _Ktv2.contract.FilterLogs(opts, "Withdrew")
	if err != nil {
		return nil, err
	}
	return &Ktv2WithdrewIterator{contract: _Ktv2.contract, event: "Withdrew", logs: logs, sub: sub}, nil
}

// WatchWithdrew is a free log subscription operation binding the contract event 0xb244b9a17ad633c6e83b7983ee04320484956a68ddbe96a0b70dfca1cf19d723.
//
// Solidity: event Withdrew(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) WatchWithdrew(opts *bind.WatchOpts, sink chan<- *Ktv2Withdrew) (event.Subscription, error) {

	logs, sub, err := _Ktv2.contract.WatchLogs(opts, "Withdrew")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2Withdrew)
				if err := _Ktv2.contract.UnpackLog(event, "Withdrew", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrew is a log parse operation binding the contract event 0xb244b9a17ad633c6e83b7983ee04320484956a68ddbe96a0b70dfca1cf19d723.
//
// Solidity: event Withdrew(address arg0, uint256 arg1)
func (_Ktv2 *Ktv2Filterer) ParseWithdrew(log types.Log) (*Ktv2Withdrew, error) {
	event := new(Ktv2Withdrew)
	if err := _Ktv2.contract.UnpackLog(event, "Withdrew", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
