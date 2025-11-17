package uarbitrum

import (
	"strings"

	"github.com/15ho/wallet-utils-go/uethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var l1GatewayRouterABI abi.ABI

func init() {
	parsedABI, err := abi.JSON(strings.NewReader(l1GatewayRouterABIJson))
	if err != nil {
		panic("parse l1 gateway router abi json" + err.Error())
	}
	l1GatewayRouterABI = parsedABI
}

var erc20ABI = uethereum.GetERC20ABI()

var outboundTransferDataArgs abi.Arguments

func init() {
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic("new abi type uint256" + err.Error())
	}
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		panic("new abi type bytes" + err.Error())
	}
	arguments := abi.Arguments{
		{Type: uint256Type},
		{Type: bytesType},
	}
	outboundTransferDataArgs = arguments
}

var l1GatewayRouterABIJson = `[
{
	"anonymous": false,
	"inputs": [
	{
		"indexed": false,
		"internalType": "address",
		"name": "newDefaultGateway",
		"type": "address"
	}
	],
	"name": "DefaultGatewayUpdated",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [
	{
		"indexed": true,
		"internalType": "address",
		"name": "l1Token",
		"type": "address"
	},
	{
		"indexed": true,
		"internalType": "address",
		"name": "gateway",
		"type": "address"
	}
	],
	"name": "GatewaySet",
	"type": "event"
},
{
	"anonymous": false,
	"inputs": [
	{
		"indexed": true,
		"internalType": "address",
		"name": "token",
		"type": "address"
	},
	{
		"indexed": true,
		"internalType": "address",
		"name": "_userFrom",
		"type": "address"
	},
	{
		"indexed": true,
		"internalType": "address",
		"name": "_userTo",
		"type": "address"
	},
	{
		"indexed": false,
		"internalType": "address",
		"name": "gateway",
		"type": "address"
	}
	],
	"name": "TransferRouted",
	"type": "event"
},
{
	"inputs": [
	{
		"internalType": "address",
		"name": "l1ERC20",
		"type": "address"
	}
	],
	"name": "calculateL2TokenAddress",
	"outputs": [
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	}
	],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [],
	"name": "counterpartGateway",
	"outputs": [
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	}
	],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [],
	"name": "defaultGateway",
	"outputs": [
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	}
	],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	},
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	},
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	},
	{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	},
	{
		"internalType": "bytes",
		"name": "",
		"type": "bytes"
	}
	],
	"name": "finalizeInboundTransfer",
	"outputs": [],
	"stateMutability": "payable",
	"type": "function"
},
{
	"inputs": [
	{
		"internalType": "address",
		"name": "_token",
		"type": "address"
	}
	],
	"name": "getGateway",
	"outputs": [
	{
		"internalType": "address",
		"name": "gateway",
		"type": "address"
	}
	],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [
	{
		"internalType": "address",
		"name": "_token",
		"type": "address"
	},
	{
		"internalType": "address",
		"name": "_from",
		"type": "address"
	},
	{
		"internalType": "address",
		"name": "_to",
		"type": "address"
	},
	{
		"internalType": "uint256",
		"name": "_amount",
		"type": "uint256"
	},
	{
		"internalType": "bytes",
		"name": "_data",
		"type": "bytes"
	}
	],
	"name": "getOutboundCalldata",
	"outputs": [
	{
		"internalType": "bytes",
		"name": "",
		"type": "bytes"
	}
	],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	}
	],
	"name": "l1TokenToGateway",
	"outputs": [
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	}
	],
	"stateMutability": "view",
	"type": "function"
},
{
	"inputs": [
	{
		"internalType": "address",
		"name": "_token",
		"type": "address"
	},
	{
		"internalType": "address",
		"name": "_to",
		"type": "address"
	},
	{
		"internalType": "uint256",
		"name": "_amount",
		"type": "uint256"
	},
	{
		"internalType": "uint256",
		"name": "_maxGas",
		"type": "uint256"
	},
	{
		"internalType": "uint256",
		"name": "_gasPriceBid",
		"type": "uint256"
	},
	{
		"internalType": "bytes",
		"name": "_data",
		"type": "bytes"
	}
	],
	"name": "outboundTransfer",
	"outputs": [
	{
		"internalType": "bytes",
		"name": "",
		"type": "bytes"
	}
	],
	"stateMutability": "payable",
	"type": "function"
},
{
	"inputs": [],
	"name": "postUpgradeInit",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [],
	"name": "router",
	"outputs": [
	{
		"internalType": "address",
		"name": "",
		"type": "address"
	}
	],
	"stateMutability": "view",
	"type": "function"
}
]`
