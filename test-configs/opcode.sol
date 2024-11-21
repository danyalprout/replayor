// SPDX-License-Identifier: MIT
pragma solidity 0.8.0;

contract TargetedOpcodes {
    mapping(uint256 => uint256) public storageSlots;
    uint256 public numCalls;

    function executeOpcode(uint256 opcode, uint256 numExecutions) public {
            if (opcode == 0) {
                executeSSTORE(numExecutions);
            } else {
                revert("Unsupported opcode");
            }
        }


    function executeSSTORE(uint256 numExecutions) public {
        // uint256 baseSlot = uint256(keccak256(abi.encodePacked(uint256(0))));
        assembly {
            let callCount := sload(numCalls.slot)
            for { let i := 0 } lt(i, numExecutions) { i := add(i, 1) } {
                // Calculate unique slot for each key within the mapping
                // in order
                let ptr := mload(0x40)
                mstore(ptr, i)
                // random
                // let ptr := keccak256(mload(0x40), add(i, callCoun
                // Compute the keccak256 hash of the memory area
                mstore(add(ptr, 0x20), storageSlots.slot)
                let slot := keccak256(ptr, 0x40)

                // Store the value at the calculated slot
                sstore(slot, slot)
            }
        }

        assembly {
            let callCount := sload(numCalls.slot)
            // Update numCalls to include numExecutions
            sstore(numCalls.slot, add(callCount, numExecutions))
        }
    }
}
