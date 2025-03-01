// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abis

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

// Ktv2factMetaData contains all meta data concerning the Ktv2fact contract.
var Ktv2factMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"created\",\"type\":\"address\"}],\"name\":\"Created\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_burnDest\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_dest\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_pool\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_ocPrcAddr\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tp\",\"type\":\"address\"}],\"name\":\"create\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"created\",\"outputs\":[{\"internalType\":\"contractKtv2\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// Ktv2factABI is the input ABI used to generate the binding from.
// Deprecated: Use Ktv2factMetaData.ABI instead.
var Ktv2factABI = Ktv2factMetaData.ABI

// Ktv2fact is an auto generated Go binding around an Ethereum contract.
type Ktv2fact struct {
	Ktv2factCaller     // Read-only binding to the contract
	Ktv2factTransactor // Write-only binding to the contract
	Ktv2factFilterer   // Log filterer for contract events
}

// Ktv2factCaller is an auto generated read-only Go binding around an Ethereum contract.
type Ktv2factCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Ktv2factTransactor is an auto generated write-only Go binding around an Ethereum contract.
type Ktv2factTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Ktv2factFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Ktv2factFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Ktv2factSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Ktv2factSession struct {
	Contract     *Ktv2fact         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Ktv2factCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Ktv2factCallerSession struct {
	Contract *Ktv2factCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// Ktv2factTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Ktv2factTransactorSession struct {
	Contract     *Ktv2factTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// Ktv2factRaw is an auto generated low-level Go binding around an Ethereum contract.
type Ktv2factRaw struct {
	Contract *Ktv2fact // Generic contract binding to access the raw methods on
}

// Ktv2factCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Ktv2factCallerRaw struct {
	Contract *Ktv2factCaller // Generic read-only contract binding to access the raw methods on
}

// Ktv2factTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Ktv2factTransactorRaw struct {
	Contract *Ktv2factTransactor // Generic write-only contract binding to access the raw methods on
}

// NewKtv2fact creates a new instance of Ktv2fact, bound to a specific deployed contract.
func NewKtv2fact(address common.Address, backend bind.ContractBackend) (*Ktv2fact, error) {
	contract, err := bindKtv2fact(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ktv2fact{Ktv2factCaller: Ktv2factCaller{contract: contract}, Ktv2factTransactor: Ktv2factTransactor{contract: contract}, Ktv2factFilterer: Ktv2factFilterer{contract: contract}}, nil
}

// NewKtv2factCaller creates a new read-only instance of Ktv2fact, bound to a specific deployed contract.
func NewKtv2factCaller(address common.Address, caller bind.ContractCaller) (*Ktv2factCaller, error) {
	contract, err := bindKtv2fact(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Ktv2factCaller{contract: contract}, nil
}

// NewKtv2factTransactor creates a new write-only instance of Ktv2fact, bound to a specific deployed contract.
func NewKtv2factTransactor(address common.Address, transactor bind.ContractTransactor) (*Ktv2factTransactor, error) {
	contract, err := bindKtv2fact(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Ktv2factTransactor{contract: contract}, nil
}

// NewKtv2factFilterer creates a new log filterer instance of Ktv2fact, bound to a specific deployed contract.
func NewKtv2factFilterer(address common.Address, filterer bind.ContractFilterer) (*Ktv2factFilterer, error) {
	contract, err := bindKtv2fact(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Ktv2factFilterer{contract: contract}, nil
}

// bindKtv2fact binds a generic wrapper to an already deployed contract.
func bindKtv2fact(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := Ktv2factMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ktv2fact *Ktv2factRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ktv2fact.Contract.Ktv2factCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ktv2fact *Ktv2factRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2fact.Contract.Ktv2factTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ktv2fact *Ktv2factRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ktv2fact.Contract.Ktv2factTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ktv2fact *Ktv2factCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ktv2fact.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ktv2fact *Ktv2factTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ktv2fact.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ktv2fact *Ktv2factTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ktv2fact.Contract.contract.Transact(opts, method, params...)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Ktv2fact *Ktv2factCaller) Count(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Ktv2fact.contract.Call(opts, &out, "count")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Ktv2fact *Ktv2factSession) Count() (*big.Int, error) {
	return _Ktv2fact.Contract.Count(&_Ktv2fact.CallOpts)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Ktv2fact *Ktv2factCallerSession) Count() (*big.Int, error) {
	return _Ktv2fact.Contract.Count(&_Ktv2fact.CallOpts)
}

// Created is a free data retrieval call binding the contract method 0x82cb6b72.
//
// Solidity: function created(uint256 ) view returns(address)
func (_Ktv2fact *Ktv2factCaller) Created(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Ktv2fact.contract.Call(opts, &out, "created", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Created is a free data retrieval call binding the contract method 0x82cb6b72.
//
// Solidity: function created(uint256 ) view returns(address)
func (_Ktv2fact *Ktv2factSession) Created(arg0 *big.Int) (common.Address, error) {
	return _Ktv2fact.Contract.Created(&_Ktv2fact.CallOpts, arg0)
}

// Created is a free data retrieval call binding the contract method 0x82cb6b72.
//
// Solidity: function created(uint256 ) view returns(address)
func (_Ktv2fact *Ktv2factCallerSession) Created(arg0 *big.Int) (common.Address, error) {
	return _Ktv2fact.Contract.Created(&_Ktv2fact.CallOpts, arg0)
}

// Create is a paid mutator transaction binding the contract method 0x43f70917.
//
// Solidity: function create(address _burnDest, address _token, address _dest, address _pool, address _ocPrcAddr, address _tp) returns()
func (_Ktv2fact *Ktv2factTransactor) Create(opts *bind.TransactOpts, _burnDest common.Address, _token common.Address, _dest common.Address, _pool common.Address, _ocPrcAddr common.Address, _tp common.Address) (*types.Transaction, error) {
	return _Ktv2fact.contract.Transact(opts, "create", _burnDest, _token, _dest, _pool, _ocPrcAddr, _tp)
}

// Create is a paid mutator transaction binding the contract method 0x43f70917.
//
// Solidity: function create(address _burnDest, address _token, address _dest, address _pool, address _ocPrcAddr, address _tp) returns()
func (_Ktv2fact *Ktv2factSession) Create(_burnDest common.Address, _token common.Address, _dest common.Address, _pool common.Address, _ocPrcAddr common.Address, _tp common.Address) (*types.Transaction, error) {
	return _Ktv2fact.Contract.Create(&_Ktv2fact.TransactOpts, _burnDest, _token, _dest, _pool, _ocPrcAddr, _tp)
}

// Create is a paid mutator transaction binding the contract method 0x43f70917.
//
// Solidity: function create(address _burnDest, address _token, address _dest, address _pool, address _ocPrcAddr, address _tp) returns()
func (_Ktv2fact *Ktv2factTransactorSession) Create(_burnDest common.Address, _token common.Address, _dest common.Address, _pool common.Address, _ocPrcAddr common.Address, _tp common.Address) (*types.Transaction, error) {
	return _Ktv2fact.Contract.Create(&_Ktv2fact.TransactOpts, _burnDest, _token, _dest, _pool, _ocPrcAddr, _tp)
}

// Ktv2factCreatedIterator is returned from FilterCreated and is used to iterate over the raw logs and unpacked data for Created events raised by the Ktv2fact contract.
type Ktv2factCreatedIterator struct {
	Event *Ktv2factCreated // Event containing the contract specifics and raw log

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
func (it *Ktv2factCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Ktv2factCreated)
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
		it.Event = new(Ktv2factCreated)
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
func (it *Ktv2factCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Ktv2factCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Ktv2factCreated represents a Created event raised by the Ktv2fact contract.
type Ktv2factCreated struct {
	Created common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterCreated is a free log retrieval operation binding the contract event 0x1449abf21e49fd025f33495e77f7b1461caefdd3d4bb646424a3f445c4576a5b.
//
// Solidity: event Created(address created)
func (_Ktv2fact *Ktv2factFilterer) FilterCreated(opts *bind.FilterOpts) (*Ktv2factCreatedIterator, error) {

	logs, sub, err := _Ktv2fact.contract.FilterLogs(opts, "Created")
	if err != nil {
		return nil, err
	}
	return &Ktv2factCreatedIterator{contract: _Ktv2fact.contract, event: "Created", logs: logs, sub: sub}, nil
}

// WatchCreated is a free log subscription operation binding the contract event 0x1449abf21e49fd025f33495e77f7b1461caefdd3d4bb646424a3f445c4576a5b.
//
// Solidity: event Created(address created)
func (_Ktv2fact *Ktv2factFilterer) WatchCreated(opts *bind.WatchOpts, sink chan<- *Ktv2factCreated) (event.Subscription, error) {

	logs, sub, err := _Ktv2fact.contract.WatchLogs(opts, "Created")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Ktv2factCreated)
				if err := _Ktv2fact.contract.UnpackLog(event, "Created", log); err != nil {
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

// ParseCreated is a log parse operation binding the contract event 0x1449abf21e49fd025f33495e77f7b1461caefdd3d4bb646424a3f445c4576a5b.
//
// Solidity: event Created(address created)
func (_Ktv2fact *Ktv2factFilterer) ParseCreated(log types.Log) (*Ktv2factCreated, error) {
	event := new(Ktv2factCreated)
	if err := _Ktv2fact.contract.UnpackLog(event, "Created", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
