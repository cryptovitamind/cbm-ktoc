// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.16;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

// MockERC20 is a freely-mintable ERC20 used only by the local integration test
// to stand in for the real staking token. Not for production.
contract MockERC20 is ERC20 {
    constructor() ERC20("MockStake", "MOCK") {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}
