// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.16;

import "@openzeppelin/contracts/access/Ownable.sol";

/// @notice The give() entry point every Burn Bank (Ktv2) exposes.
interface IBurnBank {
    function give() external payable;
}

/**
 * @title BurnBankVault
 * @notice Escrow for Burn Bank Battles challenges.
 *
 * A donor funds a vault under the keccak256 hash of a secret 32-byte key.
 * Nobody, the owner included, can withdraw that ETH. The only way out is
 * unlock(key, bank): anyone who knows the key may direct the whole deposit
 * into the give() function of a whitelisted Burn Bank, once.
 *
 * Design decisions, deliberately:
 *  - One deposit per key hash. fund() reverts if the hash is already used, so
 *    the challenge amount a site posts is exactly what the key unlocks. Top-ups
 *    need a fresh key and a fresh challenge.
 *  - No withdraw, no rescue, no expiry. A lost key means the ETH is stuck for
 *    good. That is the escrow guarantee the challenge relies on.
 *  - The unlock is a race by design. Once a key is public anyone can spend it,
 *    and a mempool watcher can front-run the unlocker with a different bank.
 *    The Battle rules treat "snatching" as fair play; a site that needs the
 *    matcher to choose the bank must reveal the key privately.
 *  - Attribution comes from this contract's VaultUnlocked event. The bank's
 *    own Gave event names this vault as the giver, not the donor.
 *  - The whitelist is checked at unlock time. Removing a bank strands nothing:
 *    the unlocker simply picks another.
 *  - Plain ETH transfers revert. With no withdraw path, ETH sent without a key
 *    hash would be lost, so receive() refuses it.
 */
contract BurnBankVault is Ownable {
    struct Vault {
        address funder;
        uint amount;
        uint fundedBlock;
        bool spent;
    }

    mapping(bytes32 => Vault) public vaults;
    mapping(address => bool) public banks;
    uint public totalLocked;

    event BankAdded(address indexed bank);
    event BankRemoved(address indexed bank);
    event VaultFunded(bytes32 indexed keyHash, address indexed funder, uint amount);
    event VaultUnlocked(bytes32 indexed keyHash, address indexed bank, uint amount, address indexed unlocker);

    constructor(address[] memory initialBanks) Ownable() {
        for (uint i = 0; i < initialBanks.length; i++) {
            _addBank(initialBanks[i]);
        }
    }

    // ------------------------------------------------------------------ owner

    function addBank(address bank) external onlyOwner {
        _addBank(bank);
    }

    function removeBank(address bank) external onlyOwner {
        require(banks[bank], "Not a bank");
        banks[bank] = false;
        emit BankRemoved(bank);
    }

    function _addBank(address bank) private {
        require(bank != address(0), "Zero bank");
        require(!banks[bank], "Already a bank");
        banks[bank] = true;
        emit BankAdded(bank);
    }

    // ----------------------------------------------------------------- public

    /// @notice Lock msg.value under keyHash = keccak256(abi.encodePacked(key)).
    function fund(bytes32 keyHash) external payable {
        require(msg.value > 0, "No ETH");
        require(vaults[keyHash].amount == 0, "Hash already funded");
        vaults[keyHash] = Vault({funder: msg.sender, amount: msg.value, fundedBlock: block.number, spent: false});
        totalLocked += msg.value;
        emit VaultFunded(keyHash, msg.sender, msg.value);
    }

    /// @notice Spend the vault for `key` by calling give() on a whitelisted bank.
    function unlock(bytes32 key, address bank) external {
        bytes32 keyHash = hashKey(key);
        Vault storage v = vaults[keyHash];
        require(v.amount > 0, "Unknown key");
        require(!v.spent, "Already unlocked");
        require(banks[bank], "Not a bank");

        // Effects before the external call: give() forwards ETH onward with a
        // low-level call, so the vault must already be marked spent.
        v.spent = true;
        uint amount = v.amount;
        totalLocked -= amount;

        IBurnBank(bank).give{value: amount}();
        emit VaultUnlocked(keyHash, bank, amount, msg.sender);
    }

    /// @notice The hash a key must be funded under.
    function hashKey(bytes32 key) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(key));
    }

    /// @notice Convenience view for sites: whether a hash is funded and unspent.
    function isUnlockable(bytes32 keyHash) external view returns (bool) {
        Vault storage v = vaults[keyHash];
        return v.amount > 0 && !v.spent;
    }

    receive() external payable {
        revert("Use fund(keyHash)");
    }
}
