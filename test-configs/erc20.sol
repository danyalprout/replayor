// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract ReplayorStressToken is ERC20, Ownable {
    uint256 public constant RATE = 10e18;
    mapping(address => uint256) balance;

    constructor() ERC20("ReplayorStressToken", "RSST") Ownable(msg.sender) {
        _mint(msg.sender, 1000_000_000 * RATE);
    }
}