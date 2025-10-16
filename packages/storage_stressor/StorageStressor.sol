// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

contract StressStorage {
    // Event with three indexed parameters to maximize log storage
    event StorageWrite(
        bytes32 indexed data, 
        bytes32 indexed slotKey, 
        bytes32 indexed value
    );

    mapping(bytes32 => bytes32) public slots;
    // Nested mapping for deeper storage
    mapping(bytes32 => mapping(bytes32 => bytes32)) public nestedSlots;
    
    // Counter to ensure unique storage writes
    uint256 public counter;
    
    // Track created contract addresses
    uint256 public contractCounter;
    address[] public createdContracts;
    
    // Function to write to multiple storage slots with pseudo-random access
    function writeStorage(uint256 slotCount) public {
        for (uint256 i = 0; i < slotCount; i++) {
            bytes32 slotKey = keccak256(abi.encode(
                block.timestamp,
                i,      // Ensures uniqueness within the loop
                counter // Ensures uniqueness across different function calls
            ));
            
            bytes32 value = keccak256(abi.encode(slotKey, i));
            
            // Write to storage with pseudo-random keys
            slots[slotKey] = value;
            
            // Write to nested mapping for deeper storage
            bytes32 nestedKey = keccak256(abi.encode(slotKey, "nested"));
            nestedSlots[slotKey][nestedKey] = value;
            
            emit StorageWrite(slotKey, nestedKey, value);
            counter++;
        }
    }

    // Deploy contracts with maximum storage slot declarations but minimal gas usage
    function deployContracts(uint256 count) public {
        // Get the current counter value for this transaction
        uint256 startCounter = contractCounter;
        
        for (uint256 i = 0; i < count; i++) {
            // Create unique salt that's different across transactions and blocks
            bytes32 salt = keccak256(abi.encode(
                msg.sender,     // Different for different callers
                block.number,   // Different for different blocks
                startCounter+i  // Different for different transactions and iterations
            ));
            
            // Deploy with CREATE2 using the unique salt
            ChildStorage child = new ChildStorage{salt: salt}();
            createdContracts.push(address(child));
        }
        
        // Update counter for next transaction
        contractCounter += count;
    }
}

contract ChildStorage {
    event ContractCreated(
        address indexed deployer,
        address indexed contractAddress,
        uint256 indexed blockNumber
    );

    // Use fixed-size arrays for different types
    // Each creates 100 sequential storage slots
    uint256[100] public uintArray;
    bytes32[100] public bytes32Array;
    address[100] public addressArray;
    bool[100] public boolArray;
    
    // Dynamic arrays - these hash to random slots
    // Each consumes a slot for length, then calculates positions via keccak
    uint256[] public dynamicArray1;
    uint256[] public dynamicArray2;
    uint256[] public dynamicArray3;
    uint256[][] public doubleArray;

    // Mappings with different key and value types create maximal
    // distribution in the storage trie
    mapping(uint256 => uint256) public simpleMap;
    mapping(address => uint256) public addressMap;
    mapping(bytes32 => bytes32) public bytes32Map;

    // Nested mappings create exponential storage combinations
    mapping(uint256 => mapping(uint256 => uint256)) public doubleMap;
    mapping(uint256 => mapping(uint256 => mapping(uint256 => uint256))) public tripleMap;
    
    // Even just declaring these without assigning stresses storage layout
    bytes public dynamicBytes;
    string public dynamicString;
    
    // Constructor with minimal execution 
    constructor() {
        // Emit an event to stress the bloom filter
        emit ContractCreated(msg.sender, address(this), block.number);
    }
}