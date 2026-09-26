// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package vault

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

// BurnBankVaultMetaData contains all meta data concerning the BurnBankVault contract.
var BurnBankVaultMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"initialBanks\",\"type\":\"address[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bank\",\"type\":\"address\"}],\"name\":\"BankAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bank\",\"type\":\"address\"}],\"name\":\"BankRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"keyHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"VaultFunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"keyHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bank\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"unlocker\",\"type\":\"address\"}],\"name\":\"VaultUnlocked\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bank\",\"type\":\"address\"}],\"name\":\"addBank\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"banks\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"keyHash\",\"type\":\"bytes32\"}],\"name\":\"fund\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"}],\"name\":\"hashKey\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"keyHash\",\"type\":\"bytes32\"}],\"name\":\"isUnlockable\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bank\",\"type\":\"address\"}],\"name\":\"removeBank\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalLocked\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"bank\",\"type\":\"address\"}],\"name\":\"unlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"vaults\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fundedBlock\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"spent\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60806040523480156200001157600080fd5b5060405162000d6938038062000d69833981016040819052620000349162000209565b6200003f3362000096565b60005b81518110156200008e5762000079828281518110620000655762000065620002db565b6020026020010151620000e660201b60201c565b806200008581620002f1565b91505062000042565b505062000319565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6001600160a01b0381166200012e5760405162461bcd60e51b81526020600482015260096024820152685a65726f2062616e6b60b81b60448201526064015b60405180910390fd5b6001600160a01b03811660009081526002602052604090205460ff16156200018a5760405162461bcd60e51b815260206004820152600e60248201526d416c726561647920612062616e6b60901b604482015260640162000125565b6001600160a01b038116600081815260026020526040808220805460ff19166001179055517f709bbc75e8831f324862b8e9f7f52f652d46c4ad81f244c9e7d0c6b51de924929190a250565b634e487b7160e01b600052604160045260246000fd5b80516001600160a01b03811681146200020457600080fd5b919050565b600060208083850312156200021d57600080fd5b82516001600160401b03808211156200023557600080fd5b818501915085601f8301126200024a57600080fd5b8151818111156200025f576200025f620001d6565b8060051b604051601f19603f83011681018181108582111715620002875762000287620001d6565b604052918252848201925083810185019188831115620002a657600080fd5b938501935b82851015620002cf57620002bf85620001ec565b84529385019392850192620002ab565b98975050505050505050565b634e487b7160e01b600052603260045260246000fd5b6000600182016200031257634e487b7160e01b600052601160045260246000fd5b5060010190565b610a4080620003296000396000f3fe6080604052600436106100ab5760003560e01c8063947223971161006457806394722397146101de5780639649650c146101fe578063bf14c1191461021e578063cca23bf214610231578063d9d84dd4146102af578063f2fde38b146102cf57600080fd5b806327978c85146100f65780634257068014610118578063568914121461014d578063715018a61461017157806380c3b8c2146101865780638da5cb5b146101b657600080fd5b366100f15760405162461bcd60e51b81526020600482015260116024820152705573652066756e64286b6579486173682960781b60448201526064015b60405180910390fd5b600080fd5b34801561010257600080fd5b50610116610111366004610968565b6102ef565b005b34801561012457600080fd5b50610138610133366004610994565b6104c4565b60405190151581526020015b60405180910390f35b34801561015957600080fd5b5061016360035481565b604051908152602001610144565b34801561017d57600080fd5b506101166104f2565b34801561019257600080fd5b506101386101a13660046109ad565b60026020526000908152604090205460ff1681565b3480156101c257600080fd5b506000546040516001600160a01b039091168152602001610144565b3480156101ea57600080fd5b506101166101f93660046109ad565b610506565b34801561020a57600080fd5b506101166102193660046109ad565b61051a565b61011661022c366004610994565b6105c0565b34801561023d57600080fd5b5061028361024c366004610994565b600160208190526000918252604090912080549181015460028201546003909201546001600160a01b039093169290919060ff1684565b604080516001600160a01b03909516855260208501939093529183015215156060820152608001610144565b3480156102bb57600080fd5b506101636102ca366004610994565b610712565b3480156102db57600080fd5b506101166102ea3660046109ad565b610744565b60006102fa83610712565b6000818152600160208190526040909120908101549192509061034d5760405162461bcd60e51b815260206004820152600b60248201526a556e6b6e6f776e206b657960a81b60448201526064016100e8565b600381015460ff16156103955760405162461bcd60e51b815260206004820152601060248201526f105b1c9958591e481d5b9b1bd8dad95960821b60448201526064016100e8565b6001600160a01b03831660009081526002602052604090205460ff166103ea5760405162461bcd60e51b815260206004820152600a6024820152694e6f7420612062616e6b60b01b60448201526064016100e8565b6003808201805460ff191660019081179091558201548154909182916000906104149084906109de565b92505081905550836001600160a01b0316639e96a23a826040518263ffffffff1660e01b81526004016000604051808303818588803b15801561045657600080fd5b505af115801561046a573d6000803e3d6000fd5b5050505050336001600160a01b0316846001600160a01b0316847fe6406b4c485bebc9610d68bd913c5f1ff5a5ccbbe9696ac9840a2b5805c0ebdc846040516104b591815260200190565b60405180910390a45050505050565b6000818152600160208190526040822090810154158015906104eb5750600381015460ff16155b9392505050565b6104fa6107ba565b6105046000610814565b565b61050e6107ba565b61051781610864565b50565b6105226107ba565b6001600160a01b03811660009081526002602052604090205460ff166105775760405162461bcd60e51b815260206004820152600a6024820152694e6f7420612062616e6b60b01b60448201526064016100e8565b6001600160a01b038116600081815260026020526040808220805460ff19169055517f6c6b75303e41e95dbe82479ed0a94ba1d8a0d5eea16a922bb68e5daeedb188679190a250565b600034116105f95760405162461bcd60e51b815260206004820152600660248201526509cde408aa8960d31b60448201526064016100e8565b600081815260016020819052604090912001541561064f5760405162461bcd60e51b815260206004820152601360248201527212185cda08185b1c9958591e48199d5b991959606a1b60448201526064016100e8565b6040805160808101825233815234602080830182815243848601908152600060608601818152888252600194859052968120955186546001600160a01b0319166001600160a01b039091161786559151928501929092559051600284015592516003928301805460ff191691151591909117905581549092906106d39084906109f7565b9091555050604051348152339082907f1fa8604660801919020e8dcfa1d3c84353f3f24643966fffbf71bd3ec3ae8e389060200160405180910390a350565b60008160405160200161072791815260200190565b604051602081830303815290604052805190602001209050919050565b61074c6107ba565b6001600160a01b0381166107b15760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016100e8565b61051781610814565b6000546001600160a01b031633146105045760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016100e8565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6001600160a01b0381166108a65760405162461bcd60e51b81526020600482015260096024820152685a65726f2062616e6b60b81b60448201526064016100e8565b6001600160a01b03811660009081526002602052604090205460ff16156109005760405162461bcd60e51b815260206004820152600e60248201526d416c726561647920612062616e6b60901b60448201526064016100e8565b6001600160a01b038116600081815260026020526040808220805460ff19166001179055517f709bbc75e8831f324862b8e9f7f52f652d46c4ad81f244c9e7d0c6b51de924929190a250565b80356001600160a01b038116811461096357600080fd5b919050565b6000806040838503121561097b57600080fd5b8235915061098b6020840161094c565b90509250929050565b6000602082840312156109a657600080fd5b5035919050565b6000602082840312156109bf57600080fd5b6104eb8261094c565b634e487b7160e01b600052601160045260246000fd5b818103818111156109f1576109f16109c8565b92915050565b808201808211156109f1576109f16109c856fea26469706673582212207a307ab6e57c7ef1cf1bd51aee742058376799898136fb389b6461af79a7a64764736f6c63430008100033",
}

// BurnBankVaultABI is the input ABI used to generate the binding from.
// Deprecated: Use BurnBankVaultMetaData.ABI instead.
var BurnBankVaultABI = BurnBankVaultMetaData.ABI

// BurnBankVaultBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BurnBankVaultMetaData.Bin instead.
var BurnBankVaultBin = BurnBankVaultMetaData.Bin

// DeployBurnBankVault deploys a new Ethereum contract, binding an instance of BurnBankVault to it.
func DeployBurnBankVault(auth *bind.TransactOpts, backend bind.ContractBackend, initialBanks []common.Address) (common.Address, *types.Transaction, *BurnBankVault, error) {
	parsed, err := BurnBankVaultMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BurnBankVaultBin), backend, initialBanks)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &BurnBankVault{BurnBankVaultCaller: BurnBankVaultCaller{contract: contract}, BurnBankVaultTransactor: BurnBankVaultTransactor{contract: contract}, BurnBankVaultFilterer: BurnBankVaultFilterer{contract: contract}}, nil
}

// BurnBankVault is an auto generated Go binding around an Ethereum contract.
type BurnBankVault struct {
	BurnBankVaultCaller     // Read-only binding to the contract
	BurnBankVaultTransactor // Write-only binding to the contract
	BurnBankVaultFilterer   // Log filterer for contract events
}

// BurnBankVaultCaller is an auto generated read-only Go binding around an Ethereum contract.
type BurnBankVaultCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BurnBankVaultTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BurnBankVaultTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BurnBankVaultFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BurnBankVaultFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BurnBankVaultSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BurnBankVaultSession struct {
	Contract     *BurnBankVault    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BurnBankVaultCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BurnBankVaultCallerSession struct {
	Contract *BurnBankVaultCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// BurnBankVaultTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BurnBankVaultTransactorSession struct {
	Contract     *BurnBankVaultTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// BurnBankVaultRaw is an auto generated low-level Go binding around an Ethereum contract.
type BurnBankVaultRaw struct {
	Contract *BurnBankVault // Generic contract binding to access the raw methods on
}

// BurnBankVaultCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BurnBankVaultCallerRaw struct {
	Contract *BurnBankVaultCaller // Generic read-only contract binding to access the raw methods on
}

// BurnBankVaultTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BurnBankVaultTransactorRaw struct {
	Contract *BurnBankVaultTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBurnBankVault creates a new instance of BurnBankVault, bound to a specific deployed contract.
func NewBurnBankVault(address common.Address, backend bind.ContractBackend) (*BurnBankVault, error) {
	contract, err := bindBurnBankVault(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BurnBankVault{BurnBankVaultCaller: BurnBankVaultCaller{contract: contract}, BurnBankVaultTransactor: BurnBankVaultTransactor{contract: contract}, BurnBankVaultFilterer: BurnBankVaultFilterer{contract: contract}}, nil
}

// NewBurnBankVaultCaller creates a new read-only instance of BurnBankVault, bound to a specific deployed contract.
func NewBurnBankVaultCaller(address common.Address, caller bind.ContractCaller) (*BurnBankVaultCaller, error) {
	contract, err := bindBurnBankVault(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BurnBankVaultCaller{contract: contract}, nil
}

// NewBurnBankVaultTransactor creates a new write-only instance of BurnBankVault, bound to a specific deployed contract.
func NewBurnBankVaultTransactor(address common.Address, transactor bind.ContractTransactor) (*BurnBankVaultTransactor, error) {
	contract, err := bindBurnBankVault(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BurnBankVaultTransactor{contract: contract}, nil
}

// NewBurnBankVaultFilterer creates a new log filterer instance of BurnBankVault, bound to a specific deployed contract.
func NewBurnBankVaultFilterer(address common.Address, filterer bind.ContractFilterer) (*BurnBankVaultFilterer, error) {
	contract, err := bindBurnBankVault(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BurnBankVaultFilterer{contract: contract}, nil
}

// bindBurnBankVault binds a generic wrapper to an already deployed contract.
func bindBurnBankVault(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BurnBankVaultMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BurnBankVault *BurnBankVaultRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BurnBankVault.Contract.BurnBankVaultCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BurnBankVault *BurnBankVaultRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BurnBankVault.Contract.BurnBankVaultTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BurnBankVault *BurnBankVaultRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BurnBankVault.Contract.BurnBankVaultTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BurnBankVault *BurnBankVaultCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BurnBankVault.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BurnBankVault *BurnBankVaultTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BurnBankVault.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BurnBankVault *BurnBankVaultTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BurnBankVault.Contract.contract.Transact(opts, method, params...)
}

// Banks is a free data retrieval call binding the contract method 0x80c3b8c2.
//
// Solidity: function banks(address ) view returns(bool)
func (_BurnBankVault *BurnBankVaultCaller) Banks(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _BurnBankVault.contract.Call(opts, &out, "banks", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Banks is a free data retrieval call binding the contract method 0x80c3b8c2.
//
// Solidity: function banks(address ) view returns(bool)
func (_BurnBankVault *BurnBankVaultSession) Banks(arg0 common.Address) (bool, error) {
	return _BurnBankVault.Contract.Banks(&_BurnBankVault.CallOpts, arg0)
}

// Banks is a free data retrieval call binding the contract method 0x80c3b8c2.
//
// Solidity: function banks(address ) view returns(bool)
func (_BurnBankVault *BurnBankVaultCallerSession) Banks(arg0 common.Address) (bool, error) {
	return _BurnBankVault.Contract.Banks(&_BurnBankVault.CallOpts, arg0)
}

// HashKey is a free data retrieval call binding the contract method 0xd9d84dd4.
//
// Solidity: function hashKey(bytes32 key) pure returns(bytes32)
func (_BurnBankVault *BurnBankVaultCaller) HashKey(opts *bind.CallOpts, key [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _BurnBankVault.contract.Call(opts, &out, "hashKey", key)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashKey is a free data retrieval call binding the contract method 0xd9d84dd4.
//
// Solidity: function hashKey(bytes32 key) pure returns(bytes32)
func (_BurnBankVault *BurnBankVaultSession) HashKey(key [32]byte) ([32]byte, error) {
	return _BurnBankVault.Contract.HashKey(&_BurnBankVault.CallOpts, key)
}

// HashKey is a free data retrieval call binding the contract method 0xd9d84dd4.
//
// Solidity: function hashKey(bytes32 key) pure returns(bytes32)
func (_BurnBankVault *BurnBankVaultCallerSession) HashKey(key [32]byte) ([32]byte, error) {
	return _BurnBankVault.Contract.HashKey(&_BurnBankVault.CallOpts, key)
}

// IsUnlockable is a free data retrieval call binding the contract method 0x42570680.
//
// Solidity: function isUnlockable(bytes32 keyHash) view returns(bool)
func (_BurnBankVault *BurnBankVaultCaller) IsUnlockable(opts *bind.CallOpts, keyHash [32]byte) (bool, error) {
	var out []interface{}
	err := _BurnBankVault.contract.Call(opts, &out, "isUnlockable", keyHash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsUnlockable is a free data retrieval call binding the contract method 0x42570680.
//
// Solidity: function isUnlockable(bytes32 keyHash) view returns(bool)
func (_BurnBankVault *BurnBankVaultSession) IsUnlockable(keyHash [32]byte) (bool, error) {
	return _BurnBankVault.Contract.IsUnlockable(&_BurnBankVault.CallOpts, keyHash)
}

// IsUnlockable is a free data retrieval call binding the contract method 0x42570680.
//
// Solidity: function isUnlockable(bytes32 keyHash) view returns(bool)
func (_BurnBankVault *BurnBankVaultCallerSession) IsUnlockable(keyHash [32]byte) (bool, error) {
	return _BurnBankVault.Contract.IsUnlockable(&_BurnBankVault.CallOpts, keyHash)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BurnBankVault *BurnBankVaultCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BurnBankVault.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BurnBankVault *BurnBankVaultSession) Owner() (common.Address, error) {
	return _BurnBankVault.Contract.Owner(&_BurnBankVault.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BurnBankVault *BurnBankVaultCallerSession) Owner() (common.Address, error) {
	return _BurnBankVault.Contract.Owner(&_BurnBankVault.CallOpts)
}

// TotalLocked is a free data retrieval call binding the contract method 0x56891412.
//
// Solidity: function totalLocked() view returns(uint256)
func (_BurnBankVault *BurnBankVaultCaller) TotalLocked(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BurnBankVault.contract.Call(opts, &out, "totalLocked")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalLocked is a free data retrieval call binding the contract method 0x56891412.
//
// Solidity: function totalLocked() view returns(uint256)
func (_BurnBankVault *BurnBankVaultSession) TotalLocked() (*big.Int, error) {
	return _BurnBankVault.Contract.TotalLocked(&_BurnBankVault.CallOpts)
}

// TotalLocked is a free data retrieval call binding the contract method 0x56891412.
//
// Solidity: function totalLocked() view returns(uint256)
func (_BurnBankVault *BurnBankVaultCallerSession) TotalLocked() (*big.Int, error) {
	return _BurnBankVault.Contract.TotalLocked(&_BurnBankVault.CallOpts)
}

// Vaults is a free data retrieval call binding the contract method 0xcca23bf2.
//
// Solidity: function vaults(bytes32 ) view returns(address funder, uint256 amount, uint256 fundedBlock, bool spent)
func (_BurnBankVault *BurnBankVaultCaller) Vaults(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Funder      common.Address
	Amount      *big.Int
	FundedBlock *big.Int
	Spent       bool
}, error) {
	var out []interface{}
	err := _BurnBankVault.contract.Call(opts, &out, "vaults", arg0)

	outstruct := new(struct {
		Funder      common.Address
		Amount      *big.Int
		FundedBlock *big.Int
		Spent       bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Funder = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.FundedBlock = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Spent = *abi.ConvertType(out[3], new(bool)).(*bool)

	return *outstruct, err

}

// Vaults is a free data retrieval call binding the contract method 0xcca23bf2.
//
// Solidity: function vaults(bytes32 ) view returns(address funder, uint256 amount, uint256 fundedBlock, bool spent)
func (_BurnBankVault *BurnBankVaultSession) Vaults(arg0 [32]byte) (struct {
	Funder      common.Address
	Amount      *big.Int
	FundedBlock *big.Int
	Spent       bool
}, error) {
	return _BurnBankVault.Contract.Vaults(&_BurnBankVault.CallOpts, arg0)
}

// Vaults is a free data retrieval call binding the contract method 0xcca23bf2.
//
// Solidity: function vaults(bytes32 ) view returns(address funder, uint256 amount, uint256 fundedBlock, bool spent)
func (_BurnBankVault *BurnBankVaultCallerSession) Vaults(arg0 [32]byte) (struct {
	Funder      common.Address
	Amount      *big.Int
	FundedBlock *big.Int
	Spent       bool
}, error) {
	return _BurnBankVault.Contract.Vaults(&_BurnBankVault.CallOpts, arg0)
}

// AddBank is a paid mutator transaction binding the contract method 0x94722397.
//
// Solidity: function addBank(address bank) returns()
func (_BurnBankVault *BurnBankVaultTransactor) AddBank(opts *bind.TransactOpts, bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.contract.Transact(opts, "addBank", bank)
}

// AddBank is a paid mutator transaction binding the contract method 0x94722397.
//
// Solidity: function addBank(address bank) returns()
func (_BurnBankVault *BurnBankVaultSession) AddBank(bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.Contract.AddBank(&_BurnBankVault.TransactOpts, bank)
}

// AddBank is a paid mutator transaction binding the contract method 0x94722397.
//
// Solidity: function addBank(address bank) returns()
func (_BurnBankVault *BurnBankVaultTransactorSession) AddBank(bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.Contract.AddBank(&_BurnBankVault.TransactOpts, bank)
}

// Fund is a paid mutator transaction binding the contract method 0xbf14c119.
//
// Solidity: function fund(bytes32 keyHash) payable returns()
func (_BurnBankVault *BurnBankVaultTransactor) Fund(opts *bind.TransactOpts, keyHash [32]byte) (*types.Transaction, error) {
	return _BurnBankVault.contract.Transact(opts, "fund", keyHash)
}

// Fund is a paid mutator transaction binding the contract method 0xbf14c119.
//
// Solidity: function fund(bytes32 keyHash) payable returns()
func (_BurnBankVault *BurnBankVaultSession) Fund(keyHash [32]byte) (*types.Transaction, error) {
	return _BurnBankVault.Contract.Fund(&_BurnBankVault.TransactOpts, keyHash)
}

// Fund is a paid mutator transaction binding the contract method 0xbf14c119.
//
// Solidity: function fund(bytes32 keyHash) payable returns()
func (_BurnBankVault *BurnBankVaultTransactorSession) Fund(keyHash [32]byte) (*types.Transaction, error) {
	return _BurnBankVault.Contract.Fund(&_BurnBankVault.TransactOpts, keyHash)
}

// RemoveBank is a paid mutator transaction binding the contract method 0x9649650c.
//
// Solidity: function removeBank(address bank) returns()
func (_BurnBankVault *BurnBankVaultTransactor) RemoveBank(opts *bind.TransactOpts, bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.contract.Transact(opts, "removeBank", bank)
}

// RemoveBank is a paid mutator transaction binding the contract method 0x9649650c.
//
// Solidity: function removeBank(address bank) returns()
func (_BurnBankVault *BurnBankVaultSession) RemoveBank(bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.Contract.RemoveBank(&_BurnBankVault.TransactOpts, bank)
}

// RemoveBank is a paid mutator transaction binding the contract method 0x9649650c.
//
// Solidity: function removeBank(address bank) returns()
func (_BurnBankVault *BurnBankVaultTransactorSession) RemoveBank(bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.Contract.RemoveBank(&_BurnBankVault.TransactOpts, bank)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BurnBankVault *BurnBankVaultTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BurnBankVault.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BurnBankVault *BurnBankVaultSession) RenounceOwnership() (*types.Transaction, error) {
	return _BurnBankVault.Contract.RenounceOwnership(&_BurnBankVault.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BurnBankVault *BurnBankVaultTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BurnBankVault.Contract.RenounceOwnership(&_BurnBankVault.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BurnBankVault *BurnBankVaultTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BurnBankVault.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BurnBankVault *BurnBankVaultSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BurnBankVault.Contract.TransferOwnership(&_BurnBankVault.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BurnBankVault *BurnBankVaultTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BurnBankVault.Contract.TransferOwnership(&_BurnBankVault.TransactOpts, newOwner)
}

// Unlock is a paid mutator transaction binding the contract method 0x27978c85.
//
// Solidity: function unlock(bytes32 key, address bank) returns()
func (_BurnBankVault *BurnBankVaultTransactor) Unlock(opts *bind.TransactOpts, key [32]byte, bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.contract.Transact(opts, "unlock", key, bank)
}

// Unlock is a paid mutator transaction binding the contract method 0x27978c85.
//
// Solidity: function unlock(bytes32 key, address bank) returns()
func (_BurnBankVault *BurnBankVaultSession) Unlock(key [32]byte, bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.Contract.Unlock(&_BurnBankVault.TransactOpts, key, bank)
}

// Unlock is a paid mutator transaction binding the contract method 0x27978c85.
//
// Solidity: function unlock(bytes32 key, address bank) returns()
func (_BurnBankVault *BurnBankVaultTransactorSession) Unlock(key [32]byte, bank common.Address) (*types.Transaction, error) {
	return _BurnBankVault.Contract.Unlock(&_BurnBankVault.TransactOpts, key, bank)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BurnBankVault *BurnBankVaultTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BurnBankVault.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BurnBankVault *BurnBankVaultSession) Receive() (*types.Transaction, error) {
	return _BurnBankVault.Contract.Receive(&_BurnBankVault.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BurnBankVault *BurnBankVaultTransactorSession) Receive() (*types.Transaction, error) {
	return _BurnBankVault.Contract.Receive(&_BurnBankVault.TransactOpts)
}

// BurnBankVaultBankAddedIterator is returned from FilterBankAdded and is used to iterate over the raw logs and unpacked data for BankAdded events raised by the BurnBankVault contract.
type BurnBankVaultBankAddedIterator struct {
	Event *BurnBankVaultBankAdded // Event containing the contract specifics and raw log

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
func (it *BurnBankVaultBankAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BurnBankVaultBankAdded)
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
		it.Event = new(BurnBankVaultBankAdded)
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
func (it *BurnBankVaultBankAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BurnBankVaultBankAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BurnBankVaultBankAdded represents a BankAdded event raised by the BurnBankVault contract.
type BurnBankVaultBankAdded struct {
	Bank common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterBankAdded is a free log retrieval operation binding the contract event 0x709bbc75e8831f324862b8e9f7f52f652d46c4ad81f244c9e7d0c6b51de92492.
//
// Solidity: event BankAdded(address indexed bank)
func (_BurnBankVault *BurnBankVaultFilterer) FilterBankAdded(opts *bind.FilterOpts, bank []common.Address) (*BurnBankVaultBankAddedIterator, error) {

	var bankRule []interface{}
	for _, bankItem := range bank {
		bankRule = append(bankRule, bankItem)
	}

	logs, sub, err := _BurnBankVault.contract.FilterLogs(opts, "BankAdded", bankRule)
	if err != nil {
		return nil, err
	}
	return &BurnBankVaultBankAddedIterator{contract: _BurnBankVault.contract, event: "BankAdded", logs: logs, sub: sub}, nil
}

// WatchBankAdded is a free log subscription operation binding the contract event 0x709bbc75e8831f324862b8e9f7f52f652d46c4ad81f244c9e7d0c6b51de92492.
//
// Solidity: event BankAdded(address indexed bank)
func (_BurnBankVault *BurnBankVaultFilterer) WatchBankAdded(opts *bind.WatchOpts, sink chan<- *BurnBankVaultBankAdded, bank []common.Address) (event.Subscription, error) {

	var bankRule []interface{}
	for _, bankItem := range bank {
		bankRule = append(bankRule, bankItem)
	}

	logs, sub, err := _BurnBankVault.contract.WatchLogs(opts, "BankAdded", bankRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BurnBankVaultBankAdded)
				if err := _BurnBankVault.contract.UnpackLog(event, "BankAdded", log); err != nil {
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

// ParseBankAdded is a log parse operation binding the contract event 0x709bbc75e8831f324862b8e9f7f52f652d46c4ad81f244c9e7d0c6b51de92492.
//
// Solidity: event BankAdded(address indexed bank)
func (_BurnBankVault *BurnBankVaultFilterer) ParseBankAdded(log types.Log) (*BurnBankVaultBankAdded, error) {
	event := new(BurnBankVaultBankAdded)
	if err := _BurnBankVault.contract.UnpackLog(event, "BankAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BurnBankVaultBankRemovedIterator is returned from FilterBankRemoved and is used to iterate over the raw logs and unpacked data for BankRemoved events raised by the BurnBankVault contract.
type BurnBankVaultBankRemovedIterator struct {
	Event *BurnBankVaultBankRemoved // Event containing the contract specifics and raw log

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
func (it *BurnBankVaultBankRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BurnBankVaultBankRemoved)
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
		it.Event = new(BurnBankVaultBankRemoved)
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
func (it *BurnBankVaultBankRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BurnBankVaultBankRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BurnBankVaultBankRemoved represents a BankRemoved event raised by the BurnBankVault contract.
type BurnBankVaultBankRemoved struct {
	Bank common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterBankRemoved is a free log retrieval operation binding the contract event 0x6c6b75303e41e95dbe82479ed0a94ba1d8a0d5eea16a922bb68e5daeedb18867.
//
// Solidity: event BankRemoved(address indexed bank)
func (_BurnBankVault *BurnBankVaultFilterer) FilterBankRemoved(opts *bind.FilterOpts, bank []common.Address) (*BurnBankVaultBankRemovedIterator, error) {

	var bankRule []interface{}
	for _, bankItem := range bank {
		bankRule = append(bankRule, bankItem)
	}

	logs, sub, err := _BurnBankVault.contract.FilterLogs(opts, "BankRemoved", bankRule)
	if err != nil {
		return nil, err
	}
	return &BurnBankVaultBankRemovedIterator{contract: _BurnBankVault.contract, event: "BankRemoved", logs: logs, sub: sub}, nil
}

// WatchBankRemoved is a free log subscription operation binding the contract event 0x6c6b75303e41e95dbe82479ed0a94ba1d8a0d5eea16a922bb68e5daeedb18867.
//
// Solidity: event BankRemoved(address indexed bank)
func (_BurnBankVault *BurnBankVaultFilterer) WatchBankRemoved(opts *bind.WatchOpts, sink chan<- *BurnBankVaultBankRemoved, bank []common.Address) (event.Subscription, error) {

	var bankRule []interface{}
	for _, bankItem := range bank {
		bankRule = append(bankRule, bankItem)
	}

	logs, sub, err := _BurnBankVault.contract.WatchLogs(opts, "BankRemoved", bankRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BurnBankVaultBankRemoved)
				if err := _BurnBankVault.contract.UnpackLog(event, "BankRemoved", log); err != nil {
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

// ParseBankRemoved is a log parse operation binding the contract event 0x6c6b75303e41e95dbe82479ed0a94ba1d8a0d5eea16a922bb68e5daeedb18867.
//
// Solidity: event BankRemoved(address indexed bank)
func (_BurnBankVault *BurnBankVaultFilterer) ParseBankRemoved(log types.Log) (*BurnBankVaultBankRemoved, error) {
	event := new(BurnBankVaultBankRemoved)
	if err := _BurnBankVault.contract.UnpackLog(event, "BankRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BurnBankVaultOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BurnBankVault contract.
type BurnBankVaultOwnershipTransferredIterator struct {
	Event *BurnBankVaultOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BurnBankVaultOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BurnBankVaultOwnershipTransferred)
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
		it.Event = new(BurnBankVaultOwnershipTransferred)
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
func (it *BurnBankVaultOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BurnBankVaultOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BurnBankVaultOwnershipTransferred represents a OwnershipTransferred event raised by the BurnBankVault contract.
type BurnBankVaultOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BurnBankVault *BurnBankVaultFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BurnBankVaultOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BurnBankVault.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BurnBankVaultOwnershipTransferredIterator{contract: _BurnBankVault.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BurnBankVault *BurnBankVaultFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BurnBankVaultOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BurnBankVault.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BurnBankVaultOwnershipTransferred)
				if err := _BurnBankVault.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BurnBankVault *BurnBankVaultFilterer) ParseOwnershipTransferred(log types.Log) (*BurnBankVaultOwnershipTransferred, error) {
	event := new(BurnBankVaultOwnershipTransferred)
	if err := _BurnBankVault.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BurnBankVaultVaultFundedIterator is returned from FilterVaultFunded and is used to iterate over the raw logs and unpacked data for VaultFunded events raised by the BurnBankVault contract.
type BurnBankVaultVaultFundedIterator struct {
	Event *BurnBankVaultVaultFunded // Event containing the contract specifics and raw log

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
func (it *BurnBankVaultVaultFundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BurnBankVaultVaultFunded)
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
		it.Event = new(BurnBankVaultVaultFunded)
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
func (it *BurnBankVaultVaultFundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BurnBankVaultVaultFundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BurnBankVaultVaultFunded represents a VaultFunded event raised by the BurnBankVault contract.
type BurnBankVaultVaultFunded struct {
	KeyHash [32]byte
	Funder  common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterVaultFunded is a free log retrieval operation binding the contract event 0x1fa8604660801919020e8dcfa1d3c84353f3f24643966fffbf71bd3ec3ae8e38.
//
// Solidity: event VaultFunded(bytes32 indexed keyHash, address indexed funder, uint256 amount)
func (_BurnBankVault *BurnBankVaultFilterer) FilterVaultFunded(opts *bind.FilterOpts, keyHash [][32]byte, funder []common.Address) (*BurnBankVaultVaultFundedIterator, error) {

	var keyHashRule []interface{}
	for _, keyHashItem := range keyHash {
		keyHashRule = append(keyHashRule, keyHashItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _BurnBankVault.contract.FilterLogs(opts, "VaultFunded", keyHashRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &BurnBankVaultVaultFundedIterator{contract: _BurnBankVault.contract, event: "VaultFunded", logs: logs, sub: sub}, nil
}

// WatchVaultFunded is a free log subscription operation binding the contract event 0x1fa8604660801919020e8dcfa1d3c84353f3f24643966fffbf71bd3ec3ae8e38.
//
// Solidity: event VaultFunded(bytes32 indexed keyHash, address indexed funder, uint256 amount)
func (_BurnBankVault *BurnBankVaultFilterer) WatchVaultFunded(opts *bind.WatchOpts, sink chan<- *BurnBankVaultVaultFunded, keyHash [][32]byte, funder []common.Address) (event.Subscription, error) {

	var keyHashRule []interface{}
	for _, keyHashItem := range keyHash {
		keyHashRule = append(keyHashRule, keyHashItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _BurnBankVault.contract.WatchLogs(opts, "VaultFunded", keyHashRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BurnBankVaultVaultFunded)
				if err := _BurnBankVault.contract.UnpackLog(event, "VaultFunded", log); err != nil {
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

// ParseVaultFunded is a log parse operation binding the contract event 0x1fa8604660801919020e8dcfa1d3c84353f3f24643966fffbf71bd3ec3ae8e38.
//
// Solidity: event VaultFunded(bytes32 indexed keyHash, address indexed funder, uint256 amount)
func (_BurnBankVault *BurnBankVaultFilterer) ParseVaultFunded(log types.Log) (*BurnBankVaultVaultFunded, error) {
	event := new(BurnBankVaultVaultFunded)
	if err := _BurnBankVault.contract.UnpackLog(event, "VaultFunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BurnBankVaultVaultUnlockedIterator is returned from FilterVaultUnlocked and is used to iterate over the raw logs and unpacked data for VaultUnlocked events raised by the BurnBankVault contract.
type BurnBankVaultVaultUnlockedIterator struct {
	Event *BurnBankVaultVaultUnlocked // Event containing the contract specifics and raw log

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
func (it *BurnBankVaultVaultUnlockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BurnBankVaultVaultUnlocked)
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
		it.Event = new(BurnBankVaultVaultUnlocked)
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
func (it *BurnBankVaultVaultUnlockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BurnBankVaultVaultUnlockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BurnBankVaultVaultUnlocked represents a VaultUnlocked event raised by the BurnBankVault contract.
type BurnBankVaultVaultUnlocked struct {
	KeyHash  [32]byte
	Bank     common.Address
	Amount   *big.Int
	Unlocker common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVaultUnlocked is a free log retrieval operation binding the contract event 0xe6406b4c485bebc9610d68bd913c5f1ff5a5ccbbe9696ac9840a2b5805c0ebdc.
//
// Solidity: event VaultUnlocked(bytes32 indexed keyHash, address indexed bank, uint256 amount, address indexed unlocker)
func (_BurnBankVault *BurnBankVaultFilterer) FilterVaultUnlocked(opts *bind.FilterOpts, keyHash [][32]byte, bank []common.Address, unlocker []common.Address) (*BurnBankVaultVaultUnlockedIterator, error) {

	var keyHashRule []interface{}
	for _, keyHashItem := range keyHash {
		keyHashRule = append(keyHashRule, keyHashItem)
	}
	var bankRule []interface{}
	for _, bankItem := range bank {
		bankRule = append(bankRule, bankItem)
	}

	var unlockerRule []interface{}
	for _, unlockerItem := range unlocker {
		unlockerRule = append(unlockerRule, unlockerItem)
	}

	logs, sub, err := _BurnBankVault.contract.FilterLogs(opts, "VaultUnlocked", keyHashRule, bankRule, unlockerRule)
	if err != nil {
		return nil, err
	}
	return &BurnBankVaultVaultUnlockedIterator{contract: _BurnBankVault.contract, event: "VaultUnlocked", logs: logs, sub: sub}, nil
}

// WatchVaultUnlocked is a free log subscription operation binding the contract event 0xe6406b4c485bebc9610d68bd913c5f1ff5a5ccbbe9696ac9840a2b5805c0ebdc.
//
// Solidity: event VaultUnlocked(bytes32 indexed keyHash, address indexed bank, uint256 amount, address indexed unlocker)
func (_BurnBankVault *BurnBankVaultFilterer) WatchVaultUnlocked(opts *bind.WatchOpts, sink chan<- *BurnBankVaultVaultUnlocked, keyHash [][32]byte, bank []common.Address, unlocker []common.Address) (event.Subscription, error) {

	var keyHashRule []interface{}
	for _, keyHashItem := range keyHash {
		keyHashRule = append(keyHashRule, keyHashItem)
	}
	var bankRule []interface{}
	for _, bankItem := range bank {
		bankRule = append(bankRule, bankItem)
	}

	var unlockerRule []interface{}
	for _, unlockerItem := range unlocker {
		unlockerRule = append(unlockerRule, unlockerItem)
	}

	logs, sub, err := _BurnBankVault.contract.WatchLogs(opts, "VaultUnlocked", keyHashRule, bankRule, unlockerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BurnBankVaultVaultUnlocked)
				if err := _BurnBankVault.contract.UnpackLog(event, "VaultUnlocked", log); err != nil {
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

// ParseVaultUnlocked is a log parse operation binding the contract event 0xe6406b4c485bebc9610d68bd913c5f1ff5a5ccbbe9696ac9840a2b5805c0ebdc.
//
// Solidity: event VaultUnlocked(bytes32 indexed keyHash, address indexed bank, uint256 amount, address indexed unlocker)
func (_BurnBankVault *BurnBankVaultFilterer) ParseVaultUnlocked(log types.Log) (*BurnBankVaultVaultUnlocked, error) {
	event := new(BurnBankVaultVaultUnlocked)
	if err := _BurnBankVault.contract.UnpackLog(event, "VaultUnlocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
