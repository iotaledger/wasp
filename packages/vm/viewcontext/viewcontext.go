package viewcontext

import (
	"math/big"
	"time"

	"go.uber.org/zap"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// ViewContext implements the needed infrastructure to run external view calls, its more lightweight than vmcontext
type ViewContext struct {
	processors            *processors.Cache
	stateReader           state.State
	chainID               isc.ChainID
	log                   *logger.Logger
	chainInfo             *isc.ChainInfo
	gasBurnLog            *gas.BurnLog
	gasBudget             uint64
	gasBurnEnabled        bool
	gasBurnLoggingEnabled bool
	callStack             []*callContext
	schemaVersion         isc.SchemaVersion
}

var _ execution.WaspCallContext = &ViewContext{}

func New(ch chain.ChainCore, stateReader state.State, gasBurnLoggingEnabled bool) (*ViewContext, error) {
	chainID := ch.ID()
	return &ViewContext{
		processors:            ch.Processors(),
		stateReader:           stateReader,
		chainID:               chainID,
		log:                   ch.Log().Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
		gasBurnLoggingEnabled: gasBurnLoggingEnabled,
		schemaVersion:         stateReader.SchemaVersion(),
	}, nil
}

func (ctx *ViewContext) stateReaderWithGasBurn() kv.KVStoreReader {
	return execution.NewKVStoreReaderWithGasBurn(ctx.stateReader, ctx)
}

func (ctx *ViewContext) contractStateReaderWithGasBurn(contract isc.Hname) kv.KVStoreReader {
	return isc.ContractStateSubrealmR(ctx.stateReaderWithGasBurn(), contract)
}

func (ctx *ViewContext) LocateProgram(programHash hashing.HashValue) (vmtype string, binary []byte, err error) {
	blobState := blob.NewStateReader(ctx.contractStateReaderWithGasBurn(blob.Contract.Hname()))
	return blobState.LocateProgram(programHash)
}

func (ctx *ViewContext) GetContractRecord(contractHname isc.Hname) (ret *root.ContractRecord) {
	rootState := root.NewStateReader(ctx.contractStateReaderWithGasBurn(root.Contract.Hname()))
	return rootState.FindContract(contractHname)
}

func (ctx *ViewContext) GasBurn(burnCode gas.BurnCode, par ...uint64) {
	if !ctx.gasBurnEnabled {
		return
	}
	g := burnCode.Cost(par...)
	ctx.gasBurnLog.Record(burnCode, g)
	if g > ctx.gasBudget {
		panic(vm.ErrGasBudgetExceeded)
	}
	ctx.gasBudget -= g
}

func (ctx *ViewContext) CurrentContractAccountID() isc.AgentID {
	hname := ctx.CurrentContractHname()
	if corecontracts.IsCoreHname(hname) {
		return accounts.CommonAccount()
	}
	return isc.NewContractAgentID(ctx.ChainID(), hname)
}

func (ctx *ViewContext) Caller() isc.AgentID {
	switch len(ctx.callStack) {
	case 0:
		panic("getCallContext: stack is empty")
	case 1:
		// first call (from webapi)
		return nil
	default:
		callerHname := ctx.callStack[len(ctx.callStack)-1].contract
		return isc.NewContractAgentID(ctx.chainID, callerHname)
	}
}

func (ctx *ViewContext) Processors() *processors.Cache {
	return ctx.processors
}

func (ctx *ViewContext) accountsStateWithGasBurn() *accounts.StateReader {
	return accounts.NewStateReader(ctx.schemaVersion, ctx.contractStateReaderWithGasBurn(accounts.Contract.Hname()))
}

func (ctx *ViewContext) GetNativeTokens(agentID isc.AgentID) isc.NativeTokens {
	return ctx.accountsStateWithGasBurn().GetNativeTokens(agentID, ctx.chainID)
}

func (ctx *ViewContext) GetAccountNFTs(agentID isc.AgentID) []iotago.NFTID {
	return ctx.accountsStateWithGasBurn().GetAccountNFTs(agentID)
}

func (ctx *ViewContext) GetNFTData(nftID iotago.NFTID) *isc.NFT {
	return ctx.accountsStateWithGasBurn().GetNFTData(nftID)
}

func (ctx *ViewContext) Timestamp() time.Time {
	return ctx.stateReader.Timestamp()
}

func (ctx *ViewContext) GetBaseTokensBalance(agentID isc.AgentID) (uint64, *big.Int) {
	return ctx.accountsStateWithGasBurn().GetBaseTokensBalance(agentID, ctx.chainID)
}

func (ctx *ViewContext) GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID sui.ObjectID) *big.Int {
	return ctx.accountsStateWithGasBurn().GetNativeTokenBalance(agentID, nativeTokenID, ctx.chainID)
}

func (ctx *ViewContext) Call(msg isc.Message, _ *isc.Assets) dict.Dict {
	ctx.log.Debugf("Call. TargetContract: %s entry point: %s", msg.Target.Contract, msg.Target.EntryPoint)
	return ctx.callView(msg)
}

func (ctx *ViewContext) ChainInfo() *isc.ChainInfo {
	return ctx.chainInfo
}

func (ctx *ViewContext) ChainID() isc.ChainID {
	return ctx.chainInfo.ChainID
}

func (ctx *ViewContext) ChainOwnerID() isc.AgentID {
	return ctx.chainInfo.ChainOwnerID
}

func (ctx *ViewContext) CurrentContractHname() isc.Hname {
	return ctx.getCallContext().contract
}

func (ctx *ViewContext) Params() *isc.Params {
	return &ctx.getCallContext().params
}

func (ctx *ViewContext) ContractStateReaderWithGasBurn() kv.KVStoreReader {
	return ctx.contractStateReaderWithGasBurn(ctx.CurrentContractHname())
}

func (ctx *ViewContext) SchemaVersion() isc.SchemaVersion {
	return ctx.schemaVersion
}

func (ctx *ViewContext) GasBudgetLeft() uint64 {
	return ctx.gasBudget
}

func (ctx *ViewContext) GasBurned() uint64 {
	// view calls start with max gas
	return ctx.chainInfo.GasLimits.MaxGasExternalViewCall - ctx.gasBudget
}

func (ctx *ViewContext) GasEstimateMode() bool {
	return false
}

func (ctx *ViewContext) Infof(format string, params ...interface{}) {
	ctx.log.Infof(format, params...)
}

func (ctx *ViewContext) Debugf(format string, params ...interface{}) {
	ctx.log.Debugf(format, params...)
}

func (ctx *ViewContext) Panicf(format string, params ...interface{}) {
	ctx.log.Panicf(format, params...)
}

// only for debugging
func (ctx *ViewContext) GasBurnLog() *gas.BurnLog {
	return ctx.gasBurnLog
}

func (ctx *ViewContext) callView(msg isc.Message) (ret dict.Dict) {
	contractRecord := ctx.GetContractRecord(msg.Target.Contract)
	if contractRecord == nil {
		panic(vm.ErrContractNotFound.Create(msg.Target.Contract))
	}
	ep := execution.GetEntryPointByProgHash(ctx, msg.Target.Contract, msg.Target.EntryPoint, contractRecord.ProgramHash)

	if !ep.IsView() {
		panic("target entrypoint is not a view")
	}

	ctx.pushCallContext(msg.Target.Contract, msg.Params)
	defer ctx.popCallContext()

	return ep.Call(sandbox.NewSandboxView(ctx))
}

func (ctx *ViewContext) initAndCallView(msg isc.Message) (ret dict.Dict) {
	ctx.chainInfo = governance.NewStateReader(ctx.contractStateReaderWithGasBurn(governance.Contract.Hname())).
		GetChainInfo(ctx.chainID)
	ctx.gasBudget = ctx.chainInfo.GasLimits.MaxGasExternalViewCall
	if ctx.gasBurnLoggingEnabled {
		ctx.gasBurnLog = gas.NewGasBurnLog()
	}
	ctx.GasBurnEnable(true)
	return ctx.callView(msg)
}

// CallViewExternal calls a view from outside the VM, for example API call
func (ctx *ViewContext) CallViewExternal(msg isc.Message) (ret dict.Dict, err error) {
	err = panicutil.CatchAllButDBError(func() {
		ret = ctx.initAndCallView(msg)
	}, ctx.log, "CallViewExternal: ")
	if err != nil {
		ret = nil
	}
	return ret, err
}

// GetMerkleProof returns proof for the key. It may also contain proof of absence of the key
func (ctx *ViewContext) GetMerkleProof(key []byte) (ret *trie.MerkleProof, err error) {
	err = panicutil.CatchAllButDBError(func() {
		ret = ctx.stateReader.GetMerkleProof(key)
	}, ctx.log, "GetMerkleProof: ")
	if err != nil {
		ret = nil
	}
	return ret, err
}

// GetBlockProof returns:
// - blockInfo record in serialized form
// - proof that the blockInfo is stored under the respective key.
// Useful for proving commitment to the past state, because blockInfo contains commitment to that block
func (ctx *ViewContext) GetBlockProof(blockIndex uint32) (blockInfo *blocklog.BlockInfo, proof *trie.MerkleProof, err error) {
	err = panicutil.CatchAllButDBError(func() {
		r := ctx.initAndCallView(blocklog.ViewGetBlockInfo.Message(&blockIndex))
		blockInfo = lo.Must(blocklog.ViewGetBlockInfo.Output2.Decode(r))
		key := blocklog.Contract.FullKey(blocklog.BlockInfoKey(blockIndex))
		proof = ctx.stateReader.GetMerkleProof(key)
	}, ctx.log, "GetMerkleProof: ")
	return
}

// GetRootCommitment calculates root commitment from state.
// A valid state must return root commitment equal to the L1Commitment from the anchor
func (ctx *ViewContext) GetRootCommitment() trie.Hash {
	return ctx.stateReader.TrieRoot()
}

// GetContractStateCommitment returns commitment to the contract's state, if possible.
// To be able to retrieve state commitment for the contract's state, the state must contain
// values of contracts hname at its nil key. Otherwise, function returns error
func (ctx *ViewContext) GetContractStateCommitment(hn isc.Hname) ([]byte, error) {
	var retC []byte
	var retErr error

	err := panicutil.CatchAllButDBError(func() {
		proof := ctx.stateReader.GetMerkleProof(hn.Bytes())
		rootC := ctx.stateReader.TrieRoot()
		retErr = proof.ValidateValue(rootC, hn.Bytes())
		if retErr != nil {
			return
		}
		_, retC = proof.MustKeyWithTerminal()
	}, ctx.log, "GetMerkleProof: ")
	if err != nil {
		return nil, err
	}
	if retErr != nil {
		return nil, retErr
	}
	return retC, nil
}

func (ctx *ViewContext) GasBurnEnable(enable bool) {
	ctx.gasBurnEnabled = enable
}

func (ctx *ViewContext) GasBurnEnabled() bool {
	return ctx.gasBurnEnabled
}
