// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
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
	return evmtypes.Signer(big.NewInt(int64(e.chainID)))
}

func (e *EVMChain) GasPerIota() (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetGasPerIota.Name, nil)
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult))
}

func (e *EVMChain) BlockNumber() (*big.Int, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetBlockNumber.Name, nil)
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evm.FieldResult))
	return bal, nil
}

func (e *EVMChain) FeeColor() (colored.Color, error) {
	feeInfo, err := e.backend.CallView(governance.Contract.Name, governance.FuncGetFeeInfo.Name, dict.Dict{
		root.ParamHname: iscp.Hn(e.contractName).Bytes(),
	})
	if err != nil {
		return colored.Color{}, err
	}
	return codec.DecodeColor(feeInfo.MustGet(governance.ParamFeeColor))
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

func (e *EVMChain) GetOnChainBalance() (colored.Balances, error) {
	agentID := iscp.NewAgentID(ledgerstate.NewED25519Address(e.backend.Signer().PublicKey), 0)
	ret, err := e.backend.CallView(accounts.Contract.Name, accounts.FuncViewBalance.Name, codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: codec.EncodeAgentID(agentID),
	}))
	if err != nil {
		return nil, err
	}
	return accounts.DecodeBalances(ret)
}

func (e *EVMChain) SendTransaction(tx *types.Transaction) error {
	feeColor, feeAmount, err := e.GasLimitFee(tx)
	if err != nil {
		return err
	}
	bal, err := e.GetOnChainBalance()
	if err != nil {
		return err
	}
	fee := colored.NewBalancesForColor(feeColor, feeAmount)
	if bal[feeColor] < feeAmount {
		// make a deposit if not enough on-chain balance to cover the fees
		err = e.backend.PostOnLedgerRequest(accounts.Contract.Name, accounts.FuncDeposit.Name, fee, nil)
		if err != nil {
			return err
		}
	}
	txdata, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	// send the Ethereum transaction
	return e.backend.PostOffLedgerRequest(e.contractName, evm.FuncSendTransaction.Name, fee, dict.Dict{
		evm.FieldTransactionData: txdata,
	})
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
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetBalance.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
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
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetCode.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
		evm.FieldAddress: address.Bytes(),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetBlockByNumber.Name, paramsWithOptionalBlockNumber(blockNumber, nil))
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
	ret, err = e.backend.CallView(e.contractName, funcName, args)
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
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetBlockByHash.Name, dict.Dict{
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
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetReceipt.Name, dict.Dict{
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

func (e *EVMChain) TransactionCount(address common.Address, blockNumberOrHash rpc.BlockNumberOrHash) (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetNonce.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
		evm.FieldAddress: address.Bytes(),
	}))
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) CallContract(args ethereum.CallMsg, blockNumberOrHash rpc.BlockNumberOrHash) ([]byte, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncCallContract.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
		evm.FieldCallMsg: evmtypes.EncodeCallMsg(args),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) EstimateGas(args ethereum.CallMsg) (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncEstimateGas.Name, dict.Dict{
		evm.FieldCallMsg: evmtypes.EncodeCallMsg(args),
	})
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) StorageAt(address common.Address, key common.Hash, blockNumberOrHash rpc.BlockNumberOrHash) ([]byte, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetStorage.Name, paramsWithOptionalBlockNumberOrHash(blockNumberOrHash, dict.Dict{
		evm.FieldAddress: address.Bytes(),
		evm.FieldKey:     key.Bytes(),
	}))
	if err != nil {
		return nil, err
	}
	return ret.MustGet(evm.FieldResult), nil
}

func (e *EVMChain) BlockTransactionCountByHash(blockHash common.Hash) (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetTransactionCountByBlockHash.Name, dict.Dict{
		evm.FieldBlockHash: blockHash.Bytes(),
	})
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) BlockTransactionCountByNumber(blockNumber *big.Int) (uint64, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetTransactionCountByBlockNumber.Name, paramsWithOptionalBlockNumber(blockNumber, nil))
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint64(ret.MustGet(evm.FieldResult), 0)
}

func (e *EVMChain) Logs(q *ethereum.FilterQuery) ([]*types.Log, error) {
	ret, err := e.backend.CallView(e.contractName, evm.FuncGetLogs.Name, dict.Dict{
		evm.FieldFilterQuery: evmtypes.EncodeFilterQuery(q),
	})
	if err != nil {
		return nil, err
	}
	return evmtypes.DecodeLogs(ret.MustGet(evm.FieldResult))
}
