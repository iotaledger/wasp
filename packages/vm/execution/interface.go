package execution

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// WaspContext defines the common functionality for vm context - both used in internal/external calls (SC execution/external view calls)
type WaspContext interface {
	LocateProgram(programHash hashing.HashValue) (vmtype string, binary []byte, err error)
	GetContractRecord(contractHname iscp.Hname) (ret *root.ContractRecord)
	GasBurn(burnCode gas.BurnCode, par ...uint64)
	Processors() *processors.Cache

	// needed for sandbox
	iscp.LogInterface
	GetAssets(agentID *iscp.AgentID) *iscp.FungibleTokens
	Timestamp() int64
	AccountID() *iscp.AgentID
	GetIotaBalance(agentID *iscp.AgentID) uint64
	GetNativeTokenBalance(agentID *iscp.AgentID, tokenID *iotago.NativeTokenID) *big.Int
	Call(contractHname, entryPoint iscp.Hname, params dict.Dict, allowance *iscp.Allowance) dict.Dict
	ChainID() *iscp.ChainID
	ChainOwnerID() *iscp.AgentID
	CurrentContractHname() iscp.Hname
	ContractCreator() *iscp.AgentID
	Params() *iscp.Params
	StateReader() kv.KVStoreReader
	GasBudgetLeft() uint64
	GetAccountNFTs(agentID *iscp.AgentID) []iotago.NFTID
	GetNFTData(nftID iotago.NFTID) iscp.NFT
	L1Params() *parameters.L1
}
