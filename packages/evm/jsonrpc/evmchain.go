// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type EVMChain struct {
	backend ChainBackend
	chainID uint16
}

func NewEVMChain(backend ChainBackend, chainID uint16) *EVMChain {
	return &EVMChain{backend, chainID}
}

func (e *EVMChain) Signer() types.Signer {
	return evmutil.Signer(big.NewInt(int64(e.chainID)))
}

func (e *EVMChain) ViewCaller() errors.ViewCaller {
	return e.backend.ISCCallView
}

func (e *EVMChain) BlockNumber() (*big.Int, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetBlockNumber.Name, nil)
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evm.FieldResult))
	return bal, nil
}

func (e *EVMChain) GasRatio() (util.Ratio32, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetGasRatio.Name, nil)
	if err != nil {
		return util.Ratio32{}, err
	}
	return codec.DecodeRatio32(ret.MustGet(evm.FieldResult))
}

func (e *EVMChain) GasFeePolicy() (*gas.GasFeePolicy, error) {
	res, err := e.backend.ISCCallView(governance.Contract.Name, governance.ViewGetFeePolicy.Name, nil)
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
	if tx.ChainId().Uint64() != uint64(e.chainID) {
		return fmt.Errorf("Chain ID mismatch")
	}
	sender, err := types.Sender(e.Signer(), tx)
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
	iscGasBudgetAffordable := gasFeePolicy.AffordableGasBudgetFromAvailableTokens(balance.Uint64())

	iscGasBudgetTx := evmtypes.EVMGasToISC(evmGas, &gasRatio)
	if iscGasBudgetAffordable < iscGasBudgetTx {
		return fmt.Errorf(
			"sender doesn't have enough L2 funds to cover tx gas budget. Balance: %v, expected: %d",
			balance.String(),
			iscGasBudgetTx/gasFeePolicy.GasPerToken,
		)
	}
	return nil
}

func paramsWithOptionalBlockNumber(blockNumber *big.Int, params dict.Dict) dict.Dict {
	ret := params
	if params == nil {
		ret = dict.Dict{}
	}
	if blockNumber != nil {
		ret.Set(evm.FieldBlockNumber, blockNumber.Bytes())
	}
	return ret
}

func paramsWithOptionalBlockNumberOrHash(blockNumberOrHash rpc.BlockNumberOrHash, params dict.Dict) dict.Dict {
	if blockNumber, ok := blockNumberOrHash.Number(); ok {
		return paramsWithOptionalBlockNumber(parseBlockNumber(blockNumber), params)
	}
	ret := params
	if params == nil {
		ret = dict.Dict{}
	}
	blockHash, _ := blockNumberOrHash.Hash()
	ret.Set(evm.FieldBlockHash, blockHash.Bytes())
	return ret
}

func (e *EVMChain) Balance(address common.Address, blockNumberOrHash rpc.BlockNumberOrHash) (*big.Int, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetBalance.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
		evm.FieldAddress: address.Bytes(),
	}))
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evm.FieldResult))
	return bal, nil
}

func (e *EVMChain) Code(address common.Address, blockNumberOrHash rpc.BlockNumberOrHash) ([]byte, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetCode.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
		evm.FieldAddress: address.Bytes(),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetBlockByNumber.Name, paramsWithOptionalBlockNumber(blockNumber, nil))
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

func (e *EVMChain) getTransactionBy(funcName string, args dict.Dict) (tx *types.Transaction, blockHash common.Hash, blockNumber, index uint64, err error) {
	var ret dict.Dict
	ret, err = e.backend.ISCCallView(evm.Contract.Name, funcName, args)
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
	return e.getTransactionBy(evm.FuncGetTransactionByHash.Name, dict.Dict{
		evm.FieldTransactionHash: hash.Bytes(),
	})
}

func (e *EVMChain) TransactionByBlockHashAndIndex(hash common.Hash, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumber, indexRet uint64, err error) {
	return e.getTransactionBy(evm.FuncGetTransactionByBlockHashAndIndex.Name, dict.Dict{
		evm.FieldBlockHash:        hash.Bytes(),
		evm.FieldTransactionIndex: codec.EncodeUint64(index),
	})
}

func (e *EVMChain) TransactionByBlockNumberAndIndex(blockNumber *big.Int, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumberRet, indexRet uint64, err error) {
	return e.getTransactionBy(evm.FuncGetTransactionByBlockNumberAndIndex.Name, paramsWithOptionalBlockNumber(blockNumber, dict.Dict{
		evm.FieldTransactionIndex: codec.EncodeUint64(index),
	}))
}

func (e *EVMChain) BlockByHash(hash common.Hash) (*types.Block, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetBlockByHash.Name, dict.Dict{
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
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetReceipt.Name, dict.Dict{
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
	params := dict.Dict{
		evm.FieldAddress: address.Bytes(),
	}
	if len(blockNumberOrHash) > 0 {
		params = paramsWithOptionalBlockNumberOrHash(blockNumberOrHash[0], params)
	}
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetNonce.Name, params)
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) CallContract(args ethereum.CallMsg, blockNumberOrHash rpc.BlockNumberOrHash) ([]byte, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncCallContract.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
		evm.FieldCallMsg: evmtypes.EncodeCallMsg(args),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) EstimateGas(callMsg ethereum.CallMsg) (uint64, error) {
	return e.backend.EVMEstimateGas(callMsg)
}

func (e *EVMChain) StorageAt(address common.Address, key common.Hash, blockNumberOrHash rpc.BlockNumberOrHash) ([]byte, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetStorage.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
		evm.FieldAddress: address.Bytes(),
		evm.FieldKey:     key.Bytes(),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) BlockTransactionCountByHash(blockHash common.Hash) (uint64, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetTransactionCountByBlockHash.Name, dict.Dict{
		evm.FieldBlockHash: blockHash.Bytes(),
	})
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) BlockTransactionCountByNumber(blockNumber *big.Int) (uint64, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetTransactionCountByBlockNumber.Name, paramsWithOptionalBlockNumber(blockNumber, nil))
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) Logs(q *ethereum.FilterQuery) ([]*types.Log, error) {
	ret, err := e.backend.ISCCallView(evm.Contract.Name, evm.FuncGetLogs.Name, dict.Dict{
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
