// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.16;

interface IVault {
    function unlock(bytes32 key, address bank) external;
}

// ReentrantBank is a test-only "Burn Bank" whose give() tries to unlock the
// same vault key again from inside the payout. A correctly ordered vault has
// already marked the key spent, so the nested unlock must revert and the
// outer call must still succeed with exactly one payout.
contract ReentrantBank {
    IVault public vault;
    bytes32 public key;
    uint public payouts;
    bool public reentryReverted;

    function arm(IVault _vault, bytes32 _key) external {
        vault = _vault;
        key = _key;
    }

    function give() external payable {
        payouts++;
        if (address(vault) != address(0)) {
            try vault.unlock(key, address(this)) {
                reentryReverted = false;
            } catch {
                reentryReverted = true;
            }
        }
    }
}
