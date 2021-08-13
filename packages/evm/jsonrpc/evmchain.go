// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

type EVMChain struct {
	backend      ChainBackend
	chainID      int
	contractName string
}

func NewEVMChain(backend ChainBackend, chainID int, contractName string) *EVMChain {
	return &EVMChain{backend, chainID, contractName}
}

func (e *EVMChain) Signer() types.Signer {
	return evm.Signer(big.NewInt(int64(e.chainID)))
}

func (e *EVMChain) GasPerIota() (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetGasPerIota.Name, nil)
	if err != nil {
		return 0, err
	}
	gpi, _, err := codec.DecodeUint64(ret.MustGet(evmchain.FieldResult))
	return gpi, err
}

func (e *EVMChain) BlockNumber() (*big.Int, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetBlockNumber.Name, nil)
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evmchain.FieldResult))
	return bal, nil
}

func (e *EVMChain) FeeColor() (colored.Color, error) {
	feeInfo, err := e.backend.CallView(root.Contract.Name, governance.FuncGetFeeInfo.Name, dict.Dict{
		root.ParamHname: evmchain.Contract.Hname().Bytes(),
	})
	if err != nil {
		return colored.Color{}, err
	}
	feeColor, _, err := codec.DecodeColor(feeInfo.MustGet(governance.ParamFeeColor))
	return feeColor, err
}

func (e *EVMChain) GasLimitFee(tx *types.Transaction) (colored.Color, uint64, error) {
	gpi, err := e.GasPerIota()
	if err != nil {
		return colored.Color{}, 0, err
	}
	feeColor, err := e.FeeColor()
	if err != nil {
		return colored.Color{}, 0, err
	}
	return feeColor, tx.Gas() / gpi, nil
}

func (e *EVMChain) SendTransaction(tx *types.Transaction) error {
	feeColor, feeAmount, err := e.GasLimitFee(tx)
	if err != nil {
		return err
	}
	fee := colored.NewBalancesForColor(feeColor, feeAmount)
	// deposit fee into sender's on-chain account
	err = e.backend.PostOnLedgerRequest(accounts.Contract.Name, accounts.FuncDeposit.Name, fee, nil)
	if err != nil {
		return err
	}
	txdata, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	// send the Ethereum transaction to the evmchain contract
	return e.backend.PostOffLedgerRequest(e.contractName, evmchain.FuncSendTransaction.Name, fee, dict.Dict{
		evmchain.FieldTransactionData: txdata,
	})
}

func paramsWithOptionalBlockNumber(blockNumber *big.Int, params dict.Dict) dict.Dict {
	if params == nil {
		return nil
	}
	if blockNumber != nil {
		params.Set(evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	return params
}

func (e *EVMChain) Balance(address common.Address, blockNumber *big.Int) (*big.Int, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetBalance.Name, paramsWithOptionalBlockNumber(blockNumber, dict.Dict{
		evmchain.FieldAddress: address.Bytes(),
	}))
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evmchain.FieldResult))
	return bal, nil
}

func (e *EVMChain) Code(address common.Address, blockNumber *big.Int) ([]byte, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetCode.Name, paramsWithOptionalBlockNumber(blockNumber, dict.Dict{
		evmchain.FieldAddress: address.Bytes(),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evmchain.FieldResult), nil
}

func (e *EVMChain) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetBlockByNumber.Name, paramsWithOptionalBlockNumber(blockNumber, nil))
	if err != nil {
		return nil, err
	}

	if !ret.MustHas(evmchain.FieldResult) {
		return nil, nil
	}

	block, err := evmchain.DecodeBlock(ret.MustGet(evmchain.FieldResult))
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (e *EVMChain) getTransactionBy(funcName string, args dict.Dict) (tx *types.Transaction, blockHash common.Hash, blockNumber, index uint64, err error) {
	var ret dict.Dict
	ret, err = e.backend.CallView(e.contractName, funcName, args)
	if err != nil {
		return
	}

	if !ret.MustHas(evmchain.FieldTransaction) {
		return
	}

	tx, err = evmchain.DecodeTransaction(ret.MustGet(evmchain.FieldTransaction))
	if err != nil {
		return
	}
	blockHash = common.BytesToHash(ret.MustGet(evmchain.FieldBlockHash))
	blockNumber, _, err = codec.DecodeUint64(ret.MustGet(evmchain.FieldBlockNumber))
	if err != nil {
		return
	}
	index, _, err = codec.DecodeUint64(ret.MustGet(evmchain.FieldTransactionIndex))
	if err != nil {
		return
	}
	return
}

func (e *EVMChain) TransactionByHash(hash common.Hash) (tx *types.Transaction, blockHash common.Hash, blockNumber, index uint64, err error) {
	return e.getTransactionBy(evmchain.FuncGetTransactionByHash.Name, dict.Dict{
		evmchain.FieldTransactionHash: hash.Bytes(),
	})
}

func (e *EVMChain) TransactionByBlockHashAndIndex(hash common.Hash, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumber, indexRet uint64, err error) {
	return e.getTransactionBy(evmchain.FuncGetTransactionByBlockHashAndIndex.Name, dict.Dict{
		evmchain.FieldBlockHash:        hash.Bytes(),
		evmchain.FieldTransactionIndex: codec.EncodeUint64(index),
	})
}

func (e *EVMChain) TransactionByBlockNumberAndIndex(blockNumber *big.Int, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumberRet, indexRet uint64, err error) {
	return e.getTransactionBy(evmchain.FuncGetTransactionByBlockNumberAndIndex.Name, paramsWithOptionalBlockNumber(blockNumber, dict.Dict{
		evmchain.FieldTransactionIndex: codec.EncodeUint64(index),
	}))
}

func (e *EVMChain) BlockByHash(hash common.Hash) (*types.Block, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetBlockByHash.Name, dict.Dict{
		evmchain.FieldBlockHash: hash.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	if !ret.MustHas(evmchain.FieldResult) {
		return nil, nil
	}

	block, err := evmchain.DecodeBlock(ret.MustGet(evmchain.FieldResult))
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (e *EVMChain) TransactionReceipt(txHash common.Hash) (*evmchain.Receipt, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetReceipt.Name, dict.Dict{
		evmchain.FieldTransactionHash: txHash.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	if !ret.MustHas(evmchain.FieldResult) {
		return nil, nil
	}

	receipt, err := evmchain.DecodeReceipt(ret.MustGet(evmchain.FieldResult))
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (e *EVMChain) TransactionCount(address common.Address, blockNumber *big.Int) (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetNonce.Name, paramsWithOptionalBlockNumber(blockNumber, dict.Dict{
		evmchain.FieldAddress: address.Bytes(),
	}))
	if err != nil {
		return 0, err
	}

	n, _, err := codec.DecodeUint64(ret.MustGet(evmchain.FieldResult))
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (e *EVMChain) CallContract(args ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncCallContract.Name, paramsWithOptionalBlockNumber(blockNumber, dict.Dict{
		evmchain.FieldCallMsg: evmchain.EncodeCallMsg(args),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evmchain.FieldResult), nil
}

func (e *EVMChain) EstimateGas(args ethereum.CallMsg) (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncEstimateGas.Name, dict.Dict{
		evmchain.FieldCallMsg: evmchain.EncodeCallMsg(args),
	})
	if err != nil {
		return 0, err
	}
	gas, _, err := codec.DecodeUint64(ret.MustGet(evmchain.FieldResult))
	return gas, err
}

func (e *EVMChain) StorageAt(address common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetStorage.Name, paramsWithOptionalBlockNumber(blockNumber, dict.Dict{
		evmchain.FieldAddress: address.Bytes(),
		evmchain.FieldKey:     key.Bytes(),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evmchain.FieldResult), nil
}

func (e *EVMChain) BlockTransactionCountByHash(blockHash common.Hash) (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetBlockTransactionCountByHash.Name, dict.Dict{
		evmchain.FieldBlockHash: blockHash.Bytes(),
	})
	if err != nil {
		return 0, err
	}
	n, _, err := codec.DecodeUint64(ret.MustGet(evmchain.FieldResult))
	return n, err
}

func (e *EVMChain) BlockTransactionCountByNumber(blockNumber *big.Int) (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetBlockTransactionCountByNumber.Name, paramsWithOptionalBlockNumber(blockNumber, nil))
	if err != nil {
		return 0, err
	}
	n, _, err := codec.DecodeUint64(ret.MustGet(evmchain.FieldResult))
	return n, err
}

func (e *EVMChain) Logs(q *ethereum.FilterQuery) ([]*types.Log, error) {
	ret, err := e.backend.CallView(e.contractName, evmchain.FuncGetLogs.Name, dict.Dict{
		evmchain.FieldFilterQuery: evmchain.EncodeFilterQuery(q),
	})
	if err != nil {
		return nil, err
	}
	return evmchain.DecodeLogs(ret.MustGet(evmchain.FieldResult))
}
