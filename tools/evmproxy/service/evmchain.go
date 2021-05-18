package service

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

type EVMChain struct {
	backend ChainBackend
}

func NewEVMChain(backend ChainBackend) *EVMChain {
	return &EVMChain{backend}
}

func (e *EVMChain) BlockNumber() (*big.Int, error) {
	ret, err := e.backend.CallView(evmchain.Interface.Name, evmchain.FuncGetBlockNumber)
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evmchain.FieldResult))
	return bal, nil
}

func (e *EVMChain) SendTransaction(tx *types.Transaction) error {
	txdata, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	return e.backend.PostRequest(evmchain.Interface.Name, evmchain.FuncSendTransaction, 1024,
		evmchain.FieldTransactionData, txdata,
	)
}

func (e *EVMChain) Balance(address common.Address, blockNumber *big.Int) (*big.Int, error) {
	params := []interface{}{
		evmchain.FieldAddress, address.Bytes(),
	}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := e.backend.CallView(evmchain.Interface.Name, evmchain.FuncGetBalance, params...)
	if err != nil {
		return nil, err
	}

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evmchain.FieldBalance))
	return bal, nil
}

func (e *EVMChain) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	params := []interface{}{}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := e.backend.CallView(evmchain.Interface.Name, evmchain.FuncGetBlockByNumber, params...)
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

func (e *EVMChain) TransactionCount(address common.Address, blockNumber *big.Int) (uint64, error) {
	params := []interface{}{
		evmchain.FieldAddress, address.Bytes(),
	}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := e.backend.CallView(evmchain.Interface.Name, evmchain.FuncGetNonce, params...)
	if err != nil {
		return 0, err
	}

	n, _, err := codec.DecodeUint64(ret.MustGet(evmchain.FieldResult))
	if err != nil {
		return 0, err
	}
	return n, nil
}
