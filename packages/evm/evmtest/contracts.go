package evmtest

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var (
	//go:embed storage.abi.json
	StorageContractABI string
	//go:embed storage.bytecode.hex
	storageContractBytecodeHex string
	StorageContractBytecode    = common.FromHex(strings.TrimSpace(storageContractBytecodeHex))

	//go:embed erc20.abi.json
	ERC20ContractABI string
	//go:embed erc20.bytecode.hex
	erc20ContractBytecodeHex string
	ERC20ContractBytecode    = common.FromHex(strings.TrimSpace(erc20ContractBytecodeHex))
)
