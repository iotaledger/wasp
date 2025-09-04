// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/iotaledger/wasp/v2/packages/evm/evmutil"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/subrealm"
	"github.com/iotaledger/wasp/v2/packages/util/panicutil"
	"github.com/iotaledger/wasp/v2/packages/vm/vmexceptions"
)

type EVMEmulator struct {
	ctx         Context
	chainConfig *params.ChainConfig
	vmConfig    vm.Config
}

type Context interface {
	State() kv.KVStore
	Timestamp() uint64
	GasLimits() GasLimits
	BlockKeepAmount() int32
	MagicContracts() map[common.Address]vm.ISCMagicContract

	TakeSnapshot() int
	RevertToSnapshot(int)

	BaseTokensDecimals() uint8
	GetBaseTokensBalance(addr common.Address) *big.Int
	AddBaseTokensBalance(addr common.Address, amount *big.Int)
	SubBaseTokensBalance(addr common.Address, amount *big.Int)

	WithoutGasBurn(f func())
}

type GasLimits struct {
	Block uint64
	Call  uint64
}

var configCache *lru.Cache[int, *params.ChainConfig]

func init() {
	var err error
	configCache, err = lru.New[int, *params.ChainConfig](100)
	if err != nil {
		panic(err)
	}
}

func getConfig(chainID int) *params.ChainConfig {
	if c, ok := configCache.Get(chainID); ok {
		return c
	}
	c := &params.ChainConfig{
		ChainID:             big.NewInt(int64(chainID)),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		Ethash:              &params.EthashConfig{},
		ShanghaiTime:        new(uint64),
		CancunTime:          new(uint64),
	}
	if !c.IsShanghai(common.Big0, 0) {
		panic("ChainConfig should report EVM version as Shanghai")
	}
	configCache.Add(chainID, c)
	return c
}

const (
	keyStateDB      = "s"
	keyBlockchainDB = "b"
)

func StateDBSubrealm(store kv.KVStore) kv.KVStore {
	return subrealm.New(store, keyStateDB)
}

func StateDBSubrealmR(store kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(store, keyStateDB)
}

func BlockchainDBSubrealm(store kv.KVStore) kv.KVStore {
	return subrealm.New(store, keyBlockchainDB)
}

func BlockchainDBSubrealmR(store kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(store, keyBlockchainDB)
}

// Init initializes the EVM state with the provided genesis allocation parameters
func Init(
	emulatorState kv.KVStore,
	chainID uint16,
	gasLimits GasLimits,
	timestamp uint64,
	alloc types.GenesisAlloc,
) {
	bdb := NewBlockchainDB(emulatorState, gasLimits.Block, BlockKeepAll)
	if bdb.Initialized() {
		panic("evm state already initialized in kvstore")
	}
	bdb.Init(chainID, timestamp)

	stateDBSubrealm := StateDBSubrealm(emulatorState)
	for addr, account := range alloc {
		CreateAccount(stateDBSubrealm, addr)
		if account.Balance != nil {
			panic("balances must be 0 at genesis")
		}
		if account.Code != nil {
			SetCode(stateDBSubrealm, addr, account.Code)
		}
		for k, v := range account.Storage {
			SetState(stateDBSubrealm, addr, k, v)
		}
		SetNonce(stateDBSubrealm, addr, account.Nonce)
	}
}

func NewEVMEmulator(ctx Context) *EVMEmulator {
	gasLimits := ctx.GasLimits()
	bdb := NewBlockchainDB(ctx.State(), gasLimits.Block, ctx.BlockKeepAmount())
	chainID := 0
	ctx.WithoutGasBurn(func() {
		if !bdb.Initialized() {
			panic("must initialize genesis block first")
		}
		chainID = int(bdb.GetChainID())
	})

	return &EVMEmulator{
		ctx:         ctx,
		chainConfig: getConfig(chainID),
		vmConfig: vm.Config{
			MagicContracts: ctx.MagicContracts(),
			NoBaseFee:      true, // gas fee is set by ISC
		},
	}
}

func (e *EVMEmulator) StateDB() *StateDB {
	return NewStateDB(e.ctx)
}

func (e *EVMEmulator) BlockchainDB() *BlockchainDB {
	return NewBlockchainDB(e.ctx.State(), e.ctx.GasLimits().Block, e.ctx.BlockKeepAmount())
}

func (e *EVMEmulator) BlockGasLimit() uint64 {
	return e.ctx.GasLimits().Block
}

func (e *EVMEmulator) CallGasLimit() uint64 {
	return e.ctx.GasLimits().Call
}

func (e *EVMEmulator) ChainContext() core.ChainContext {
	return &chainContext{
		emulator: e,
		engine:   ethash.NewFaker(),
	}
}

func coreMsgFromCallMsg(call ethereum.CallMsg, gasEstimateMode bool, statedb *StateDB) *core.Message {
	return &core.Message{
		To:               call.To,
		From:             call.From,
		Nonce:            statedb.GetNonce(call.From),
		Value:            call.Value,
		GasLimit:         call.Gas,
		GasPrice:         call.GasPrice,
		GasFeeCap:        call.GasFeeCap,
		GasTipCap:        call.GasTipCap,
		Data:             call.Data,
		AccessList:       call.AccessList,
		SkipNonceChecks:  gasEstimateMode,
		SkipFromEOACheck: gasEstimateMode,
	}
}

// CallContract executes a contract call, without committing changes to the state
func (e *EVMEmulator) CallContract(call ethereum.CallMsg, gasEstimateMode bool) (*core.ExecutionResult, error) {
	// Ensure message is initialized properly.
	if call.Gas == 0 {
		call.Gas = e.ctx.GasLimits().Call
	}
	if call.Value == nil {
		call.Value = big.NewInt(0)
	}

	pendingHeader := e.BlockchainDB().GetPendingHeader(e.ctx.Timestamp())

	statedb := e.StateDB()

	// don't commit changes to state
	i := statedb.Snapshot()
	defer statedb.RevertToSnapshot(i)

	return e.applyMessage(
		coreMsgFromCallMsg(call, gasEstimateMode, statedb),
		statedb,
		pendingHeader,
		nil,
		nil,
	)
}

func (e *EVMEmulator) applyMessage(
	msg *core.Message,
	statedb vm.StateDB,
	header *types.Header,
	tracer *tracing.Hooks,
	onTxStart func(vmEnv *vm.EVM),
) (res *core.ExecutionResult, err error) {
	// Set msg gas price to 0
	msg.GasPrice = big.NewInt(0)
	msg.GasFeeCap = big.NewInt(0)
	msg.GasTipCap = big.NewInt(0)

	blockContext := core.NewEVMBlockContext(header, e.ChainContext(), nil)
	blockContext.BaseFee = new(big.Int)

	vmConfig := e.vmConfig
	vmConfig.Tracer = tracer

	vmEnv := vm.NewEVM(blockContext, statedb, e.chainConfig, vmConfig)

	if msg.GasLimit > e.ctx.GasLimits().Call {
		msg.GasLimit = e.ctx.GasLimits().Call
	}

	gasPool := core.GasPool(msg.GasLimit)
	vmEnv.SetTxContext(core.NewEVMTxContext(msg))

	if onTxStart != nil {
		onTxStart(vmEnv)
	}
	// catch any exceptions during the execution, so that an EVM receipt is always produced
	caughtErr := panicutil.CatchAllExcept(func() {
		res, err = core.ApplyMessage(vmEnv, msg, &gasPool)
	}, vmexceptions.SkipRequestErrors...)
	if caughtErr != nil {
		return &core.ExecutionResult{
			Err:        vm.ErrExecutionReverted,
			UsedGas:    msg.GasLimit - gasPool.Gas(),
			ReturnData: abiEncodeError(caughtErr),
		}, caughtErr
	}
	return res, err
}

// see UnpackRevert in go-ethereum/accounts/abi/abi.go
var revertSelector = crypto.Keccak256([]byte("Error(string)"))[:4]

func abiEncodeError(err error) []byte {
	// include the ISC error as the revert reason by encoding it into the returnData
	ret := bytes.Clone(revertSelector)
	abiString, err2 := abi.NewType("string", "", nil)
	if err2 != nil {
		panic(err2)
	}
	encodedErr, err2 := abi.Arguments{{Type: abiString}}.Pack(err.Error())
	if err2 != nil {
		panic(err2)
	}
	return append(ret, encodedErr...)
}

func (e *EVMEmulator) SendTransaction(
	tx *types.Transaction,
	tracer *tracing.Hooks,
	addToBlockchain ...bool,
) (receipt *types.Receipt, result *core.ExecutionResult, err error) {
	statedbImpl := e.StateDB()
	var statedb vm.StateDB = statedbImpl
	if tracer != nil {
		statedb = NewHookedState(statedbImpl, tracer)
	}
	pendingHeader := e.BlockchainDB().GetPendingHeader(e.ctx.Timestamp())

	sender, err := types.Sender(e.Signer(), tx)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid transaction: %w", err)
	}
	nonce := e.StateDB().GetNonce(sender)
	if tx.Nonce() != nonce {
		return nil, nil, fmt.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), nonce)
	}

	signer := types.MakeSigner(e.chainConfig, pendingHeader.Number, pendingHeader.Time)
	msg, err := core.TransactionToMessage(tx, signer, pendingHeader.BaseFee)
	if err != nil {
		return nil, nil, err
	}

	onTxStart := func(vmEnv *vm.EVM) {
		if tracer != nil && tracer.OnTxStart != nil {
			tracer.OnTxStart(vmEnv.GetVMContext(), tx, msg.From)
		}
	}

	result, err = e.applyMessage(
		msg,
		statedb,
		pendingHeader,
		tracer,
		onTxStart,
	)

	gasUsed := uint64(0)
	if result != nil {
		gasUsed = result.UsedGas
	}
	cumulativeGasUsed := e.BlockchainDB().GetPendingCumulativeGasUsed() + gasUsed

	receipt = &types.Receipt{
		Type:              tx.Type(),
		CumulativeGasUsed: cumulativeGasUsed,
		GasUsed:           gasUsed,
		Logs:              statedbImpl.GetLogs(),
	}
	receipt.Bloom = types.CreateBloom(receipt)

	if result == nil || result.Failed() {
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}

	if msg.To == nil {
		receipt.ContractAddress = crypto.CreateAddress(msg.From, tx.Nonce())
	}

	if tracer != nil && tracer.OnTxEnd != nil {
		tracer.OnTxEnd(receipt, err)
	}

	// add the tx and receipt to the blockchain unless addToBlockchain == false
	if len(addToBlockchain) == 0 || addToBlockchain[0] {
		e.BlockchainDB().AddTransaction(tx, receipt)
	}

	return receipt, result, err
}

func (e *EVMEmulator) MintBlock() {
	e.BlockchainDB().MintBlock(e.ctx.Timestamp())
}

func (e *EVMEmulator) Signer() types.Signer {
	return evmutil.Signer(e.chainConfig.ChainID)
}

type chainContext struct {
	emulator *EVMEmulator
	engine   consensus.Engine
}

var _ core.ChainContext = &chainContext{}

func (c *chainContext) Engine() consensus.Engine {
	return c.engine
}

func (c *chainContext) GetHeader(common.Hash, uint64) *types.Header {
	panic("not implemented")
}

func (c *chainContext) Config() *params.ChainConfig {
	return c.emulator.chainConfig
}
