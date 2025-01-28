// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title PrecompileTargeter
/// @notice A contract designed to target specific EVM precompiles for stress testing
/// @dev Each function maintains a gasBuffer to ensure block space for required deposit 
///      transactions so that the block gasTarget is not exceeded. Amount burned will
///      be set by the tx's gasLimit param.
contract PrecompileTargeter {

    // Precompile address: 0x01
    function ecrecoverPrecompile() public view {
        uint256 gasBuffer = 300_000;

        // Initialize with dummy signature components
        bytes32 hash = bytes32(uint256(1));
        uint8 v = 27;
        bytes32 r = bytes32(uint256(2));
        bytes32 s = bytes32(uint256(3));

        while (gasleft() > gasBuffer) {
            bytes memory input = new bytes(128);
            assembly {
                mstore(add(input, 0x20), hash)
                mstore(add(input, 0x40), v)
                mstore(add(input, 0x60), r)
                mstore(add(input, 0x80), s)
                pop(staticcall(not(0), 0x01, add(input, 0x20), 128, 0, 0))
            }

            // Modify inputs for next iteration
            hash = bytes32(uint256(hash) + 1);
            r = bytes32(uint256(r) + 1);
            s = bytes32(uint256(s) + 1);
        }
    }

    // Precompile address: 0x02
    function sha256Precompile() public view {
        uint256 gasBuffer = 300_000;

        // Initialize with some data to hash
        bytes memory data = new bytes(64);
        uint256 value = 1;

        while (gasleft() > gasBuffer) {
            assembly {
                mstore(add(data, 0x20), value)
                pop(staticcall(not(0), 0x02, add(data, 0x20), 64, 0, 0))
            }

            // Modify input for next iteration
            value++;
        }
    }

    // Precompile address: 0x03
    function ripemd160Precompile() public view {
        uint256 gasBuffer = 300_000;

        // Initialize with some data to hash
        bytes memory data = new bytes(64);
        uint256 value = 1;

        while (gasleft() > gasBuffer) {
            assembly {
                mstore(add(data, 0x20), value)
                pop(staticcall(not(0), 0x03, add(data, 0x20), 64, 0, 0))
            }

            // Modify input for next iteration
            value++;
        }
    }

    // Precompile address: 0x04
    function identityPrecompile() public view {
        uint256 gasBuffer = 300_000;

        // Initialize with some data to copy
        bytes memory data = new bytes(64);
        uint256 value = 1;

        while (gasleft() > gasBuffer) {
            assembly {
                mstore(add(data, 0x20), value)
                pop(staticcall(not(0), 0x04, add(data, 0x20), 64, 0, 0))
            }

            // Modify input for next iteration
            value++;
        }
    }

    // Precompile address: 0x05
    function modexpPrecompile() public view { 
        uint256 gasBuffer = 300_000;

        uint256 base = 2**128;
        uint256 exp = 2**129;
        uint256 mod = 2**253;

        while (gasleft() > gasBuffer) {
            uint256[3] memory input;
            input[0] = base;
            input[1] = exp;
            input[2] = mod;

            assembly {
                // Ignore return value by calling pop
                pop(staticcall(not(0), 0x05, input, 0x60, 0, 0))
            }

            // Increment the inputs for the next iteration to evade
            // caching by the execution client
            base++;
            exp++;
            mod++;
        }
    }

    // Precompile address: 0x06
    function ecaddPrecompile() public view {
        uint256 gasBuffer = 300_000;

        // Initialize points on the alt_bn128 curve
        uint256 x1 = 1;
        uint256 y1 = 2;
        uint256 x2 = 3;
        uint256 y2 = 4;

        while (gasleft() > gasBuffer) {
            uint256[4] memory input = [x1, y1, x2, y2];

            assembly {
                pop(staticcall(not(0), 0x06, input, 0x80, 0, 0))
            }

            // Increment the inputs to avoid potential client caching
            x1 = x1 + 1;
            y1 = y1 + 2;
            x2 = x2 << 1;
            y2 = y2 << 2;
        }
    }

    // Precompile address: 0x07
    function ecmulPrecompile() public view {
        uint256 gasBuffer = 300_000;

        // Initialize point on the alt_bn128 curve and scalar
        uint256 x = 1;
        uint256 y = 2;
        uint256 scalar = 3;

        while (gasleft() > gasBuffer) {
            uint256[3] memory input = [x, y, scalar];

            assembly {
                pop(staticcall(not(0), 0x07, input, 0x60, 0, 0))
            }

            // Increment the inputs to avoid potential client caching
            x = x + 1;
            y = y + 2;
            scalar = scalar + 3;
        }
    }

    // Precompile address: 0x08
    function ecpairingPrecompile() public view {
        uint256 gasBuffer = 300_000;

        // Initialize with minimal valid pairing input (two pairs)
        uint256[12] memory input;
        input[0] = 1;  // First point, x coordinate
        input[1] = 2;  // First point, y coordinate
        input[2] = 3;  // Second point, x coordinate (imaginary part)
        input[3] = 4;  // Second point, y coordinate (imaginary part)
        input[4] = 5;  // Third point, x coordinate
        input[5] = 6;  // Third point, y coordinate

        while (gasleft() > gasBuffer) {
            assembly {
                pop(staticcall(not(0), 0x08, input, 0x180, 0, 0))
            }

            // Increment all inputs to avoid potential client caching
            for (uint i = 0; i < 12; i++) {
                input[i]++;
            }
        }
    }

    // Precompile address: 0x09
    function blake2fPrecompile() public view {
        uint256 gasBuffer = 300_000;

        // Initialize with dummy blake2f input
        bytes memory input = new bytes(213);
        uint32 rounds = 12;
        assembly {
            mstore(add(input, 0x20), rounds)
        }

        while (gasleft() > gasBuffer) {
            assembly {
                pop(staticcall(not(0), 0x09, add(input, 0x20), 213, 0, 0))
            }

            // Increment rounds to avoid potential client caching
            rounds++;
            assembly {
                mstore(add(input, 0x20), rounds)
            }
        }
    }
}