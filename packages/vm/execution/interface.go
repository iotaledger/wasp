package execution

import (
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// The following interfaces define the common functionality for SC execution (VM/external view calls)

type WaspContext interface {
	LocateProgram(programHash hashing.HashValue) (vmtype string, binary []byte, err error)
	GetContractRecord(contractHname isc.Hname) (ret *root.ContractRecord)
	Processors() *processors.Cache
}

type GasContext interface {
	GasBurnEnabled() bool
	GasBurnEnable(enable bool)
	GasBurn(burnCode gas.BurnCode, par ...uint64)
	GasEstimateMode() bool
}

type WaspCallContext interface {
	WaspContext
	GasContext
	isc.LogInterface
	Timestamp() time.Time
	CurrentContractAccountID() isc.AgentID
	Caller() isc.AgentID
	GetNativeTokens(agentID isc.AgentID) iotago.NativeTokens
	GetBaseTokensBalance(agentID isc.AgentID) uint64
	GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID iotago.NativeTokenID) *big.Int
	Call(contractHname, entryPoint isc.Hname, params dict.Dict, allowance *isc.Assets) dict.Dict
	ChainID() isc.ChainID
	ChainOwnerID() isc.AgentID
	ChainInfo() *isc.ChainInfo
	CurrentContractHname() isc.Hname
	Params() *isc.Params
	ContractStateReaderWithGasBurn() kv.KVStoreReader
	GasBurned() uint64
	GasBudgetLeft() uint64
	GetAccountNFTs(agentID isc.AgentID) []iotago.NFTID
	GetNFTData(nftID iotago.NFTID) *isc.NFT
}
