package abi

import (
	"github.com/ethereum/go-ethereum/common"
)

// StandardERC20ABI is the standard ERC20 ABI containing Transfer and Approval events.
// This is used for default ERC20 event detection.
const StandardERC20ABI = `[
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "from",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "to",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "value",
        "type": "uint256"
      }
    ],
    "name": "Transfer",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "owner",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "spender",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "value",
        "type": "uint256"
      }
    ],
    "name": "Approval",
    "type": "event"
  }
]`

// RegisterStandardERC20ABI registers the standard ERC20 ABI for a contract address.
// This enables decoding of Transfer and Approval events for the given contract.
func RegisterStandardERC20ABI(registry *Registry, address string) error {
	loader := NewLoader(registry)
	return loader.LoadAndRegisterString(
		common.HexToAddress(address),
		StandardERC20ABI,
	)
}

// LoadStandardERC20ABI loads the standard ERC20 ABI.
func LoadStandardERC20ABI() (*Registry, error) {
	registry := NewRegistry()
	loader := NewLoader(registry)
	_, err := loader.LoadFromString(StandardERC20ABI)
	if err != nil {
		return nil, err
	}
	return registry, nil
}
