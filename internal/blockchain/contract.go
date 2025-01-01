package blockchain


// Contract ABI 常量，需要根据你的智能合约生成
const ContractABI = `[
    {
      "type": "constructor",
      "stateMutability": "undefined",
      "payable": false,
      "inputs": []
    },
    {
      "type": "error",
      "name": "OwnableInvalidOwner",
      "inputs": [
        {
          "type": "address",
          "name": "owner"
        }
      ]
    },
    {
      "type": "error",
      "name": "OwnableUnauthorizedAccount",
      "inputs": [
        {
          "type": "address",
          "name": "account"
        }
      ]
    },
    {
      "type": "error",
      "name": "ReentrancyGuardReentrantCall",
      "inputs": []
    },
    {
      "type": "event",
      "anonymous": false,
      "name": "BatchStatusChanged",
      "inputs": [
        {
          "type": "bytes32",
          "name": "lshTreeRoot",
          "indexed": true
        },
        {
          "type": "uint8",
          "name": "newStatus",
          "indexed": false
        }
      ]
    },
    {
      "type": "event",
      "anonymous": false,
      "name": "BatchStored",
      "inputs": [
        {
          "type": "bytes32",
          "name": "lshTreeRoot",
          "indexed": true
        }
      ]
    },
    {
      "type": "event",
      "anonymous": false,
      "name": "NodeRegistered",
      "inputs": [
        {
          "type": "address",
          "name": "nodeAddress",
          "indexed": true
        },
        {
          "type": "string",
          "name": "ipAddress",
          "indexed": false
        }
      ]
    },
    {
      "type": "event",
      "anonymous": false,
      "name": "NodeUpdated",
      "inputs": [
        {
          "type": "address",
          "name": "nodeAddress",
          "indexed": true
        }
      ]
    },
    {
      "type": "event",
      "anonymous": false,
      "name": "OwnershipTransferred",
      "inputs": [
        {
          "type": "address",
          "name": "previousOwner",
          "indexed": true
        },
        {
          "type": "address",
          "name": "newOwner",
          "indexed": true
        }
      ]
    },
    {
      "type": "function",
      "name": "EARTH_RADIUS",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "uint256",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "MAX_BATCH_SIZE",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "uint256",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "MAX_NODES_PER_BATCH",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "uint256",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "MAX_VALID_DISTANCE",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "uint256",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "SCORE_WEIGHT_CAPACITY",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "uint256",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "SCORE_WEIGHT_CREDIT",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "uint256",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "SCORE_WEIGHT_DISTANCE",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "uint256",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "batches",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [
        {
          "type": "bytes32",
          "name": ""
        }
      ],
      "outputs": [
        {
          "type": "uint256",
          "name": "timestamp"
        },
        {
          "type": "uint256",
          "name": "ttl"
        },
        {
          "type": "uint8",
          "name": "status"
        },
        {
          "type": "tuple",
          "name": "location",
          "components": [
            {
              "type": "int256",
              "name": "latitude"
            },
            {
              "type": "int256",
              "name": "longitude"
            }
          ]
        }
      ]
    },
    {
      "type": "function",
      "name": "minCreditScore",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "uint256",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "owner",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [],
      "outputs": [
        {
          "type": "address",
          "name": ""
        }
      ]
    },
    {
      "type": "function",
      "name": "queryBatches",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [
        {
          "type": "bytes32[]",
          "name": "lshTreeRoots"
        }
      ],
      "outputs": [
        {
          "type": "address[][]",
          "name": "nodesArray"
        },
        {
          "type": "string[][]",
          "name": "ipAddressesArray"
        },
        {
          "type": "bool[]",
          "name": "validBatches"
        }
      ]
    },
    {
      "type": "function",
      "name": "registerNode",
      "constant": false,
      "payable": false,
      "inputs": [
        {
          "type": "string",
          "name": "ipAddress"
        },
        {
          "type": "int256",
          "name": "latitude"
        },
        {
          "type": "int256",
          "name": "longitude"
        },
        {
          "type": "uint256",
          "name": "capacity"
        }
      ],
      "outputs": []
    },
    {
      "type": "function",
      "name": "renounceOwnership",
      "constant": false,
      "payable": false,
      "inputs": [],
      "outputs": []
    },
    {
      "type": "function",
      "name": "setMinCreditScore",
      "constant": false,
      "payable": false,
      "inputs": [
        {
          "type": "uint256",
          "name": "newScore"
        }
      ],
      "outputs": []
    },
    {
      "type": "function",
      "name": "storageNodes",
      "constant": true,
      "stateMutability": "view",
      "payable": false,
      "inputs": [
        {
          "type": "address",
          "name": ""
        }
      ],
      "outputs": [
        {
          "type": "string",
          "name": "ipAddress"
        },
        {
          "type": "uint256",
          "name": "creditScore"
        },
        {
          "type": "uint256",
          "name": "capacity"
        },
        {
          "type": "uint256",
          "name": "usedCapacity"
        },
        {
          "type": "bool",
          "name": "isActive"
        },
        {
          "type": "uint256",
          "name": "lastUpdateTime"
        },
        {
          "type": "tuple",
          "name": "location",
          "components": [
            {
              "type": "int256",
              "name": "latitude"
            },
            {
              "type": "int256",
              "name": "longitude"
            }
          ]
        }
      ]
    },
    {
      "type": "function",
      "name": "storeBatches",
      "constant": false,
      "payable": false,
      "inputs": [
        {
          "type": "bytes32[]",
          "name": "lshTreeRoots"
        },
        {
          "type": "uint256",
          "name": "ttl"
        },
        {
          "type": "int256",
          "name": "latitude"
        },
        {
          "type": "int256",
          "name": "longitude"
        }
      ],
      "outputs": [
        {
          "type": "address[]",
          "name": "selectedNodes"
        },
        {
          "type": "string[]",
          "name": "nodeIPs"
        },
        {
          "type": "uint256[]",
          "name": "nodeCreditScores"
        },
        {
          "type": "uint256[]",
          "name": "nodeCapacities"
        },
        {
          "type": "uint256[]",
          "name": "nodeUsedCapacities"
        },
        {
          "type": "int256[]",
          "name": "nodeLatitudes"
        },
        {
          "type": "int256[]",
          "name": "nodeLongitudes"
        }
      ]
    },
    {
      "type": "function",
      "name": "transferOwnership",
      "constant": false,
      "payable": false,
      "inputs": [
        {
          "type": "address",
          "name": "newOwner"
        }
      ],
      "outputs": []
    },
    {
      "type": "function",
      "name": "updateBatchStatus",
      "constant": false,
      "payable": false,
      "inputs": [
        {
          "type": "bytes32",
          "name": "lshTreeRoot"
        },
        {
          "type": "uint8",
          "name": "newStatus"
        }
      ],
      "outputs": []
    },
    {
      "type": "function",
      "name": "updateNodeCapacity",
      "constant": false,
      "payable": false,
      "inputs": [
        {
          "type": "uint256",
          "name": "newUsedCapacity"
        },
        {
          "type": "uint256",
          "name": "newTotalCapacity"
        }
      ],
      "outputs": []
    },
    {
      "type": "function",
      "name": "updateNodeCreditScore",
      "constant": false,
      "payable": false,
      "inputs": [
        {
          "type": "address",
          "name": "node"
        },
        {
          "type": "uint256",
          "name": "newScore"
        }
      ],
      "outputs": []
    },
    {
      "type": "function",
      "name": "updateNodeLocation",
      "constant": false,
      "payable": false,
      "inputs": [
        {
          "type": "int256",
          "name": "latitude"
        },
        {
          "type": "int256",
          "name": "longitude"
        }
      ],
      "outputs": []
    }
  ]`
