// Package evmnames provides the names of EVM core contract functions and other
// constants.
// It is separated from the evm interface to avoid import loops (the names are used
// in the isc package.
package evmnames

const (
	Contract = "evm"

	FuncSendTransaction   = "sendTransaction"
	FuncCallContract      = "callContract"
	ViewGetChainID        = "getChainID"
	FuncRegisterERC20Coin = "registerERC20Coin"
	FuncNewL1Deposit      = "newL1Deposit"
)
