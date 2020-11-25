package vmtypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// Sandbox is an interface given to the processor to access the VMContext
// and virtual state, transaction builder and request parameters through it.
type Sandbox interface {
	DeployContract(vmtype string, programBinary []byte, name string, description string, initParams codec.ImmutableCodec) error
	Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error)
	// only calls view entry points
	//// FIXME no need for two calls. Refactor
	//CallView(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error)

	// general
	ChainID() coretypes.ChainID
	ChainOwnerID() coretypes.AgentID
	IsRequestContext() bool
	RequestID() coretypes.RequestID

	// call context
	Params() codec.ImmutableCodec
	Caller() coretypes.AgentID
	MyContractID() coretypes.ContractID
	MyAgentID() coretypes.AgentID

	GetTimestamp() int64
	GetEntropy() hashing.HashValue // 32 bytes of deterministic and unpredictably random data

	// Same as panic(), but added as a Sandbox method to emphasize that it's ok to panic from a SC.
	// A panic will be recovered, and Rollback() will be automatically called after.
	Panic(v interface{})

	// clear all updates, restore same context as in the beginning of the VM call
	Rollback()

	// base level of virtual state access
	State() codec.MutableMustCodec
	// new implementation
	Accounts() Accounts

	// TransferToAddress send tokens to ledger address (not contract)
	TransferToAddress(addr address.Address, transfer coretypes.ColoredBalances) bool
	// TransferCrossChain send funds to the targetAgentID account cross chain
	// to move own funds to own account use MyAgentID() as a targetAgentID
	TransferCrossChain(targetAgentID coretypes.AgentID, targetChainID coretypes.ChainID, transfer coretypes.ColoredBalances) bool

	// PostRequest sends cross chain request
	PostRequest(par NewRequestParams) bool
	// PostRequestToSelf send cross chain request to the caller contract on the same chain
	PostRequestToSelf(entryPoint coretypes.Hname, args dict.Dict) bool
	// PostRequestToSelfWithDelay sends request to itself with timelock for some seconds after the current timestamp
	PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, deferForSec uint32) bool

	// for testing
	// Event and Eventf publish "vmmsg" message through Publisher on nanomsg
	Event(msg string)
	Eventf(format string, args ...interface{})
}

type NewRequestParams struct {
	TargetContractID coretypes.ContractID
	EntryPoint       coretypes.Hname
	Timelock         uint32
	Params           dict.Dict
	Transfer         coretypes.ColoredBalances
}

// Accounts is an interface to access all functions with tokens
// in the local context of the call to a smart contract
type Accounts interface {
	MyBalances() coretypes.ColoredBalances
	Incoming() coretypes.ColoredBalances
	Balance(col balance.Color) int64
	MoveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool
}
