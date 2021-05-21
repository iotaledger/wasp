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
	//deployed bytecode and runtime bytecode are different, see: https://ethereum.stackexchange.com/questions/13086/whats-the-difference-between-solcs-bin-bytecode-versus-bin-runtime
	//go:embed erc20.bytecode-runtime.hex
	ERC20ContractRuntimeBytecodeHex string
	ERC20ContractRuntimeBytecode    = common.FromHex(strings.TrimSpace(ERC20ContractRuntimeBytecodeHex))

	//go:embed loop.abi.json
	LoopContractABI string
	//go:embed loop.bytecode.hex
	loopContractBytecodeHex string
	LoopContractBytecode    = common.FromHex(strings.TrimSpace(loopContractBytecodeHex))
)
