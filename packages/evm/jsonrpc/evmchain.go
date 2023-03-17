// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	vmerrors "github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type EVMChain struct {
	backend  ChainBackend
	chainID  uint16 // cache
	newBlock *event.Event1[*NewBlockEvent]
}

type NewBlockEvent struct {
	block *types.Block
	logs  []*types.Log
}

func NewEVMChain(backend ChainBackend, pub *publisher.Publisher, log *logger.Logger) *EVMChain {
	e := &EVMChain{
		backend:  backend,
		newBlock: event.New1[*NewBlockEvent](),
	}

	pub.Events.NewBlock.Hook(func(ev *publisher.ISCEvent[*blocklog.BlockInfo]) {
		if !ev.ChainID.Equals(*e.backend.ISCChainID()) {
			return
		}
		state, err := e.backend.ISCStateByBlockIndex(ev.Payload.BlockIndex)
		if err != nil {
			log.Error(err)
			return
		}
		blockNumber := new(big.Int).SetUint64(evmBlockNumberByISCBlockIndex(ev.Payload.BlockIndex))
		block, err := e.blockByNumber(state, blockNumber)
		if err != nil {
			log.Error(err)
			return
		}

		q := &ethereum.FilterQuery{
			FromBlock: blockNumber,
			ToBlock:   blockNumber,
		}
		ret, err := e.backend.ISCCallView(state, evm.Contract.Name, evm.FuncGetLogs.Name, dict.Dict{
			evm.FieldFilterQuery: evmtypes.EncodeFilterQuery(q),
		})
		if err != nil {
			log.Error(err)
			return
		}
		logs, err := evmtypes.DecodeLogs(ret.MustGet(evm.FieldResult))
		if err != nil {
			log.Error(err)
			return
		}
		e.newBlock.Trigger(&NewBlockEvent{
			block: block,
			logs:  logs,
		})
	})

	return e
}

func (e *EVMChain) Signer() (types.Signer, error) {
	chainID, err := e.ChainID()
	if err != nil {
		return nil, err
	}
	return evmutil.Signer(big.NewInt(int64(chainID))), nil
}

func (e *EVMChain) ChainID() (uint16, error) {
	if e.chainID == 0 {
		ret, err := e.backend.ISCCallView(e.backend.ISCLatestState(), evm.Contract.Name, evm.FuncGetChainID.Name, nil)
		if err != nil {
			return 0, err
		}
		e.chainID = codec.MustDecodeUint16(ret.MustGet(evm.FieldResult))
	}
	return e.chainID, nil
}

func (e *EVMChain) ViewCaller(chainState state.State) vmerrors.ViewCaller {
	return func(contractName string, funcName string, params dict.Dict) (dict.Dict, error) {
		return e.backend.ISCCallView(chainState, contractName, funcName, params)
	}
}

func (e *EVMChain) BlockNumber() (*big.Int, error) {
	ret, err := e.backend.ISCCallView(e.backend.ISCLatestState(), evm.Contract.Name, evm.FuncGetBlockNumber.Name, nil)
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evm.FieldResult))
	return bal, nil
}

func (e *EVMChain) GasRatio() (util.Ratio32, error) {
	ret, err := e.backend.ISCCallView(
		e.backend.ISCLatestState(),
		governance.Contract.Name,
		governance.ViewGetEVMGasRatio.Name,
		nil,
	)
	if err != nil {
		return util.Ratio32{}, err
	}
	return codec.DecodeRatio32(ret.MustGet(governance.ParamEVMGasRatio))
}

func (e *EVMChain) GasFeePolicy() (*gas.FeePolicy, error) {
	res, err := e.backend.ISCCallView(e.backend.ISCLatestState(), governance.Contract.Name, governance.ViewGetFeePolicy.Name, nil)
	if err != nil {
		return nil, err
	}
	fpBin := res.MustGet(governance.ParamFeePolicyBytes)
	feePolicy, err := gas.FeePolicyFromBytes(fpBin)
	if err != nil {
		return nil, err
	}
	return feePolicy, nil
}

func (e *EVMChain) SendTransaction(tx *types.Transaction) error {
	chainID, err := e.ChainID()
	if err != nil {
		return err
	}
	if tx.ChainId().Uint64() != uint64(chainID) {
		return errors.New("chain ID mismatch")
	}
	signer, err := e.Signer()
	if err != nil {
		return err
	}
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	expectedNonce, err := e.TransactionCount(sender)
	if err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}
	if tx.Nonce() != expectedNonce {
		return fmt.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), expectedNonce)
	}

	if err := e.checkEnoughL2FundsForGasBudget(sender, tx.Gas()); err != nil {
		return err
	}
	return e.backend.EVMSendTransaction(tx)
}

func (e *EVMChain) checkEnoughL2FundsForGasBudget(sender common.Address, evmGas uint64) error {
	gasRatio, err := e.GasRatio()
	if err != nil {
		return fmt.Errorf("could not fetch gas ratio: %w", err)
	}
	balance, err := e.Balance(sender, rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber))
	if err != nil {
		return fmt.Errorf("could not fetch sender balance: %w", err)
	}
	gasFeePolicy, err := e.GasFeePolicy()
	if err != nil {
		return fmt.Errorf("could not fetch the gas fee policy: %w", err)
	}
	iscGasBudgetAffordable := gasFeePolicy.GasBudgetFromTokens(balance.Uint64())

	iscGasBudgetTx := gas.EVMGasToISC(evmGas, &gasRatio)
	if iscGasBudgetAffordable < iscGasBudgetTx {
		return fmt.Errorf(
			"sender doesn't have enough L2 funds to cover tx gas budget. Balance: %v, expected: %d",
			balance.String(),
			gasFeePolicy.FeeFromGas(iscGasBudgetTx),
		)
	}
	return nil
}

func (e *EVMChain) iscStateFromEVMBlockNumber(blockNumber *big.Int) (state.State, error) {
	if blockNumber == nil {
		return e.backend.ISCLatestState(), nil
	}
	if !blockNumber.IsUint64() {
		return nil, fmt.Errorf("block number is too large: %s", blockNumber)
	}
	iscBlockIndex, err := iscBlockIndexByEVMBlockNumber(blockNumber.Uint64())
	if err != nil {
		return nil, err
	}
	return e.backend.ISCStateByBlockIndex(iscBlockIndex)
}

func (e *EVMChain) iscStateFromEVMBlockNumberOrHash(blockNumberOrHash rpc.BlockNumberOrHash) (state.State, error) {
	if blockNumber, ok := blockNumberOrHash.Number(); ok {
		return e.iscStateFromEVMBlockNumber(parseBlockNumber(blockNumber))
	}
	blockHash, _ := blockNumberOrHash.Hash()
	block, err := e.BlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	return e.iscStateFromEVMBlockNumber(block.Number())
}

func (e *EVMChain) Balance(address common.Address, blockNumberOrHash rpc.BlockNumberOrHash) (*big.Int, error) {
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	ret, err := e.backend.ISCCallView(chainState, evm.Contract.Name, evm.FuncGetBalance.Name, dict.Dict{
		evm.FieldAddress: address.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evm.FieldResult))
	return bal, nil
}

func (e *EVMChain) Code(address common.Address, blockNumberOrHash rpc.BlockNumberOrHash) ([]byte, error) {
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	ret, err := e.backend.ISCCallView(chainState, evm.Contract.Name, evm.FuncGetCode.Name, dict.Dict{
		evm.FieldAddress: address.Bytes(),
	})
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	chainState, err := e.iscStateFromEVMBlockNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	return e.blockByNumber(chainState, blockNumber)
}

func (e *EVMChain) blockByNumber(chainState state.State, blockNumber *big.Int) (*types.Block, error) {
	params := dict.Dict{}
	if blockNumber != nil {
		params[evm.FieldBlockNumber] = blockNumber.Bytes()
	}
	ret, err := e.backend.ISCCallView(chainState, evm.Contract.Name, evm.FuncGetBlockByNumber.Name, params)
	if err != nil {
		return nil, err
	}

	if !ret.MustHas(evm.FieldResult) {
		return nil, nil
	}

	block, err := evmtypes.DecodeBlock(ret.MustGet(evm.FieldResult))
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (e *EVMChain) getTransactionBy(chainState state.State, funcName string, args dict.Dict) (tx *types.Transaction, blockHash common.Hash, blockNumber, index uint64, err error) {
	var ret dict.Dict
	ret, err = e.backend.ISCCallView(chainState, evm.Contract.Name, funcName, args)
	if err != nil {
		return
	}

	if !ret.MustHas(evm.FieldTransaction) {
		return
	}

	tx, err = evmtypes.DecodeTransaction(ret.MustGet(evm.FieldTransaction))
	if err != nil {
		return
	}
	blockHash = common.BytesToHash(ret.MustGet(evm.FieldBlockHash))
	blockNumber, err = codec.DecodeUint64(ret.MustGet(evm.FieldBlockNumber), 0)
	if err != nil {
		return
	}
	index, err = codec.DecodeUint64(ret.MustGet(evm.FieldTransactionIndex), 0)
	return
}

func (e *EVMChain) TransactionByHash(hash common.Hash) (tx *types.Transaction, blockHash common.Hash, blockNumber, index uint64, err error) {
	return e.getTransactionBy(e.backend.ISCLatestState(), evm.FuncGetTransactionByHash.Name, dict.Dict{
		evm.FieldTransactionHash: hash.Bytes(),
	})
}

func (e *EVMChain) TransactionByBlockHashAndIndex(hash common.Hash, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumber, indexRet uint64, err error) {
	return e.getTransactionBy(e.backend.ISCLatestState(), evm.FuncGetTransactionByBlockHashAndIndex.Name, dict.Dict{
		evm.FieldBlockHash:        hash.Bytes(),
		evm.FieldTransactionIndex: codec.EncodeUint64(index),
	})
}

func (e *EVMChain) TransactionByBlockNumberAndIndex(blockNumber *big.Int, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumberRet, indexRet uint64, err error) {
	blockIndex, err := e.iscStateFromEVMBlockNumber(blockNumber)
	if err != nil {
		return nil, common.Hash{}, 0, 0, err
	}
	return e.getTransactionBy(blockIndex, evm.FuncGetTransactionByBlockNumberAndIndex.Name, dict.Dict{
		evm.FieldTransactionIndex: codec.EncodeUint64(index),
	})
}

func (e *EVMChain) BlockByHash(hash common.Hash) (*types.Block, error) {
	ret, err := e.backend.ISCCallView(e.backend.ISCLatestState(), evm.Contract.Name, evm.FuncGetBlockByHash.Name, dict.Dict{
		evm.FieldBlockHash: hash.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	if !ret.MustHas(evm.FieldResult) {
		return nil, nil
	}

	block, err := evmtypes.DecodeBlock(ret.MustGet(evm.FieldResult))
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (e *EVMChain) TransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	ret, err := e.backend.ISCCallView(e.backend.ISCLatestState(), evm.Contract.Name, evm.FuncGetReceipt.Name, dict.Dict{
		evm.FieldTransactionHash: txHash.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	if !ret.MustHas(evm.FieldResult) {
		return nil, nil
	}

	receipt, err := evmtypes.DecodeReceiptFull(ret.MustGet(evm.FieldResult))
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (e *EVMChain) TransactionCount(address common.Address, blockNumberOrHash ...rpc.BlockNumberOrHash) (uint64, error) {
	var chainState state.State
	if len(blockNumberOrHash) > 0 {
		var err error
		chainState, err = e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash[0])
		if err != nil {
			return 0, err
		}
	} else {
		chainState = e.backend.ISCLatestState()
	}
	ret, err := e.backend.ISCCallView(chainState, evm.Contract.Name, evm.FuncGetNonce.Name, dict.Dict{
		evm.FieldAddress: address.Bytes(),
	})
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) CallContract(args ethereum.CallMsg, blockNumberOrHash rpc.BlockNumberOrHash) ([]byte, error) {
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	ret, err := e.backend.ISCCallView(chainState, evm.Contract.Name, evm.FuncCallContract.Name, dict.Dict{
		evm.FieldCallMsg: evmtypes.EncodeCallMsg(args),
	})
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) EstimateGas(callMsg ethereum.CallMsg) (uint64, error) {
	return e.backend.EVMEstimateGas(callMsg)
}

func (e *EVMChain) GasPrice() *big.Int {
	return e.backend.EVMGasPrice()
}

func (e *EVMChain) StorageAt(address common.Address, key common.Hash, blockNumberOrHash rpc.BlockNumberOrHash) ([]byte, error) {
	latestState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	ret, err := e.backend.ISCCallView(latestState, evm.Contract.Name, evm.FuncGetStorage.Name, dict.Dict{
		evm.FieldAddress: address.Bytes(),
		evm.FieldKey:     key.Bytes(),
	})
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) BlockTransactionCountByHash(blockHash common.Hash) (uint64, error) {
	ret, err := e.backend.ISCCallView(e.backend.ISCLatestState(), evm.Contract.Name, evm.FuncGetTransactionCountByBlockHash.Name, dict.Dict{
		evm.FieldBlockHash: blockHash.Bytes(),
	})
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) BlockTransactionCountByNumber(blockNumber *big.Int) (uint64, error) {
	chainState, err := e.iscStateFromEVMBlockNumber(blockNumber)
	if err != nil {
		return 0, err
	}
	ret, err := e.backend.ISCCallView(chainState, evm.Contract.Name, evm.FuncGetTransactionCountByBlockNumber.Name, nil)
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) Logs(q *ethereum.FilterQuery) ([]*types.Log, error) {
	ret, err := e.backend.ISCCallView(e.backend.ISCLatestState(), evm.Contract.Name, evm.FuncGetLogs.Name, dict.Dict{
		evm.FieldFilterQuery: evmtypes.EncodeFilterQuery(q),
	})
	if err != nil {
		return nil, err
	}
	return evmtypes.DecodeLogs(ret.MustGet(evm.FieldResult))
}

func (e *EVMChain) BaseToken() *parameters.BaseToken {
	return e.backend.BaseToken()
}

func (e *EVMChain) SubscribeNewHeads(ch chan<- *types.Header) (unsubscribe func()) {
	return e.newBlock.Hook(func(ev *NewBlockEvent) {
		ch <- ev.block.Header()
	}).Unhook
}

func (e *EVMChain) SubscribeLogs(q *ethereum.FilterQuery, ch chan<- []*types.Log) (unsubscribe func()) {
	return e.newBlock.Hook(func(ev *NewBlockEvent) {
		if q.BlockHash != nil && *q.BlockHash != ev.block.Hash() {
			return
		}
		if q.FromBlock != nil && q.FromBlock.IsUint64() && q.FromBlock.Cmp(ev.block.Number()) > 0 {
			return
		}
		if q.ToBlock != nil && q.ToBlock.IsUint64() && q.ToBlock.Cmp(ev.block.Number()) < 0 {
			return
		}

		var matchedLogs []*types.Log
		for _, log := range ev.logs {
			if evmtypes.LogMatches(log, q.Addresses, q.Topics) {
				matchedLogs = append(matchedLogs, log)
			}
		}
		if len(matchedLogs) > 0 {
			fmt.Printf("\n\n4\n\n")
			ch <- matchedLogs
		}
	}).Unhook
}

// the first EVM block (number 0) is "minted" at ISC block index 1 (init chain)
func iscBlockIndexByEVMBlockNumber(n uint64) (uint32, error) {
	if n > math.MaxUint32-1 {
		return 0, fmt.Errorf("block number is too large: %d", n)
	}
	return uint32(n), nil
}

// the first EVM block (number 0) is "minted" at ISC block index 1 (init chain)
func evmBlockNumberByISCBlockIndex(n uint32) uint64 {
	return uint64(n)
}
