// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.16;

// StubTokenPrice implements the TPI interface Ktv2 expects. On a fresh dev
// chain there is no Uniswap liquidity, so the real price oracle would revert
// (Ktv2's constructor requires price > 0). This returns a fixed positive price
// so the contract deploys and stakes; price never affects winner selection.
contract StubTokenPrice {
    uint256 private constant FIXED_PRICE = 1e18;

    function price(address) external pure returns (uint256) {
        return FIXED_PRICE;
    }

    function priceV2(address) external pure returns (uint256) {
        return FIXED_PRICE;
    }
}
