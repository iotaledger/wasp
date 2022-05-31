// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package jsonrpc implements JSON-RPC endpoints according to
// https://eth.wiki/json-rpc/API
package jsonrpc

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"golang.org/x/crypto/sha3"
	"golang.org/x/xerrors"
)

type EthService struct {
	evmChain *EVMChain
	accounts *AccountManager
}

func NewEthService(evmChain *EVMChain, accounts *AccountManager) *EthService {
	return &EthService{evmChain, accounts}
}

func (e *EthService) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(eth.ETH66)
}

func (e *EthService) resolveError(err error) error {
	if err == nil {
		return nil
	}
	if vmError, ok := err.(*iscp.UnresolvedVMError); ok {
		resolvedErr, resolveErr := errors.Resolve(vmError, e.evmChain.ViewCaller())
		if resolveErr != nil {
			return xerrors.Errorf("could not resolve VMError %w: %v", vmError, resolveErr)
		}
		return resolvedErr.AsGoError()
	}
	return err
}

func (e *EthService) GetTransactionCount(address common.Address, blockNumberOrHash rpc.BlockNumberOrHash) (hexutil.Uint64, error) {
	n, err := e.evmChain.TransactionCount(address, blockNumberOrHash)
	if err != nil {
		return 0, e.resolveError(err)
	}
	return hexutil.Uint64(n), nil
}

func (e *EthService) BlockNumber() (*hexutil.Big, error) {
	n, err := e.evmChain.BlockNumber()
	if err != nil {
		return nil, e.resolveError(err)
	}
	return (*hexutil.Big)(n), nil
}

func (e *EthService) GetBlockByNumber(blockNumber rpc.BlockNumber, full bool) (map[string]interface{}, error) {
	block, err := e.evmChain.BlockByNumber(parseBlockNumber(blockNumber))
	if err != nil {
		return nil, e.resolveError(err)
	}
	if block == nil {
		return nil, nil
	}
	return RPCMarshalBlock(block, true, full)
}

func (e *EthService) GetBlockByHash(hash common.Hash, full bool) (map[string]interface{}, error) {
	block, err := e.evmChain.BlockByHash(hash)
	if err != nil {
		return nil, e.resolveError(err)
	}
	if block == nil {
		return nil, nil
	}
	return RPCMarshalBlock(block, true, full)
}

func (e *EthService) GetTransactionByHash(hash common.Hash) (*RPCTransaction, error) {
	tx, blockHash, blockNumber, index, err := e.evmChain.TransactionByHash(hash)
	if err != nil {
		return nil, e.resolveError(err)
	}
	if tx == nil {
		return nil, nil
	}
	return newRPCTransaction(tx, blockHash, blockNumber, index), err
}

func (e *EthService) GetTransactionByBlockHashAndIndex(blockHash common.Hash, index hexutil.Uint) (*RPCTransaction, error) {
	tx, _, blockNumber, _, err := e.evmChain.TransactionByBlockHashAndIndex(blockHash, uint64(index))
	if err != nil {
		return nil, e.resolveError(err)
	}
	if tx == nil {
		return nil, nil
	}
	return newRPCTransaction(tx, blockHash, blockNumber, uint64(index)), err
}

func (e *EthService) GetTransactionByBlockNumberAndIndex(blockNumberOrTag rpc.BlockNumber, index hexutil.Uint) (*RPCTransaction, error) {
	tx, blockHash, blockNumber, _, err := e.evmChain.TransactionByBlockNumberAndIndex(parseBlockNumber(blockNumberOrTag), uint64(index))
	if err != nil {
		return nil, e.resolveError(err)
	}
	if tx == nil {
		return nil, nil
	}
	return newRPCTransaction(tx, blockHash, blockNumber, uint64(index)), err
}

func (e *EthService) GetBalance(address common.Address, blockNumberOrHash rpc.BlockNumberOrHash) (*hexutil.Big, error) {
	bal, err := e.evmChain.Balance(address, blockNumberOrHash)
	if err != nil {
		return nil, e.resolveError(err)
	}
	return (*hexutil.Big)(bal), nil
}

func (e *EthService) GetCode(address common.Address, blockNumberOrHash rpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	code, err := e.evmChain.Code(address, blockNumberOrHash)
	if err != nil {
		return nil, e.resolveError(err)
	}
	return hexutil.Bytes(code), nil
}

func (e *EthService) GetTransactionReceipt(txHash common.Hash) (map[string]interface{}, error) {
	r, err := e.evmChain.TransactionReceipt(txHash)
	if err != nil {
		return nil, e.resolveError(err)
	}
	if r == nil {
		return nil, nil
	}
	tx, _, _, _, err := e.evmChain.TransactionByHash(txHash) // nolint:dogsled
	if err != nil {
		return nil, e.resolveError(err)
	}
	return RPCMarshalReceipt(r, tx), nil
}

func (e *EthService) SendRawTransaction(txBytes hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(txBytes, tx); err != nil {
		return common.Hash{}, err
	}
	if err := e.evmChain.SendTransaction(tx); err != nil {
		return common.Hash{}, e.resolveError(err)
	}
	return tx.Hash(), nil
}

func (e *EthService) Call(args *RPCCallArgs, blockNumberOrHash rpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	ret, err := e.evmChain.CallContract(args.parse(), blockNumberOrHash)
	return hexutil.Bytes(ret), e.resolveError(err)
}

func (e *EthService) EstimateGas(args *RPCCallArgs) (hexutil.Uint64, error) {
	gas, err := e.evmChain.EstimateGas(args.parse())
	return hexutil.Uint64(gas), e.resolveError(err)
}

func (e *EthService) GetStorageAt(address common.Address, key common.Hash, blockNumberOrHash rpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	ret, err := e.evmChain.StorageAt(address, key, blockNumberOrHash)
	return hexutil.Bytes(ret), e.resolveError(err)
}

func (e *EthService) GetBlockTransactionCountByHash(blockHash common.Hash) (hexutil.Uint, error) {
	ret, err := e.evmChain.BlockTransactionCountByHash(blockHash)
	return hexutil.Uint(ret), e.resolveError(err)
}

func (e *EthService) GetBlockTransactionCountByNumber(blockNumber rpc.BlockNumber) (hexutil.Uint, error) {
	ret, err := e.evmChain.BlockTransactionCountByNumber(parseBlockNumber(blockNumber))
	return hexutil.Uint(ret), e.resolveError(err)
}

func (e *EthService) GetUncleCountByBlockHash(blockHash common.Hash) hexutil.Uint {
	return hexutil.Uint(0) // no uncles are ever generated
}

func (e *EthService) GetUncleCountByBlockNumber(blockNumber rpc.BlockNumber) hexutil.Uint {
	return hexutil.Uint(0) // no uncles are ever generated
}

func (e *EthService) GetUncleByBlockHashAndIndex(blockHash common.Hash, index hexutil.Uint) map[string]interface{} {
	return nil // no uncles are ever generated
}

func (e *EthService) GetUncleByBlockNumberAndIndex(blockNumberOrTag rpc.BlockNumber, index hexutil.Uint) map[string]interface{} {
	return nil // no uncles are ever generated
}

func (e *EthService) Accounts() []common.Address {
	return e.accounts.Addresses()
}

func (e *EthService) GasPrice() *hexutil.Big {
	return (*hexutil.Big)(big.NewInt(0))
}

func (e *EthService) Mining() bool {
	return false
}

func (e *EthService) Hashrate() float64 {
	return 0
}

func (e *EthService) Coinbase() common.Address {
	return common.Address{}
}

func (e *EthService) Syncing() bool {
	return false
}

func (e *EthService) GetCompilers() []string {
	return []string{}
}

func (e *EthService) Sign(addr common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	account := e.accounts.Get(addr)
	if account == nil {
		return nil, xerrors.New("Account is not unlocked")
	}

	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), string(data))
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	hash := hasher.Sum(nil)

	signed, err := crypto.Sign(hash, account)
	if err == nil {
		signed[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	}
	return signed, err
}

func (e *EthService) SignTransaction(args *SendTxArgs) (hexutil.Bytes, error) {
	tx, err := e.parseTxArgs(args)
	if err != nil {
		return nil, err
	}
	data, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(data), nil
}

func (e *EthService) SendTransaction(args *SendTxArgs) (common.Hash, error) {
	tx, err := e.parseTxArgs(args)
	if err != nil {
		return common.Hash{}, err
	}
	if err := e.evmChain.SendTransaction(tx); err != nil {
		return common.Hash{}, e.resolveError(err)
	}
	return tx.Hash(), nil
}

func (e *EthService) parseTxArgs(args *SendTxArgs) (*types.Transaction, error) {
	account := e.accounts.Get(args.From)
	if account == nil {
		return nil, xerrors.New("Account is not unlocked")
	}
	if err := args.setDefaults(e); err != nil {
		return nil, err
	}
	return types.SignTx(args.toTransaction(), e.evmChain.Signer(), account)
}

func (e *EthService) GetLogs(q *RPCFilterQuery) ([]*types.Log, error) {
	logs, err := e.evmChain.Logs((*ethereum.FilterQuery)(q))
	if err != nil {
		return nil, e.resolveError(err)
	}
	return logs, nil
}

// ChainID implements the eth_chainId method according to https://eips.ethereum.org/EIPS/eip-695
func (e *EthService) ChainId() hexutil.Uint { //nolint:revive
	return hexutil.Uint(e.evmChain.chainID)
}

/*
Not implemented:
func (e *EthService) NewFilter()
func (e *EthService) NewBlockFilter()
func (e *EthService) NewPendingTransactionFilter()
func (e *EthService) UninstallFilter()
func (e *EthService) GetFilterChanges()
func (e *EthService) GetFilterLogs()
func (e *EthService) SubmitWork()
func (e *EthService) GetWork()
func (e *EthService) SubmitHashrate()
func (e *EthService) CompileLLL()
func (e *EthService) CompileSolidity()
func (e *EthService) CompileSerpent()
*/

type NetService struct {
	chainID int
}

func NewNetService(chainID int) *NetService {
	return &NetService{chainID}
}

func (s *NetService) Version() string {
	return strconv.Itoa(s.chainID)
}

func (s *NetService) Listening() bool         { return true }
func (s *NetService) PeerCount() hexutil.Uint { return 0 }

type Web3Service struct{}

func NewWeb3Service() *Web3Service {
	return &Web3Service{}
}

func (s *Web3Service) ClientVersion() string {
	return "wasp/evmproxy"
}

func (s *Web3Service) Sha3(input hexutil.Bytes) hexutil.Bytes {
	return crypto.Keccak256(input)
}

type TxPoolService struct{}

func NewTxPoolService() *TxPoolService {
	return &TxPoolService{}
}

func (s *TxPoolService) Content() map[string]map[string]map[string]*RPCTransaction {
	return map[string]map[string]map[string]*RPCTransaction{
		"pending": make(map[string]map[string]*RPCTransaction),
		"queued":  make(map[string]map[string]*RPCTransaction),
	}
}

func (s *TxPoolService) Inspect() map[string]map[string]map[string]string {
	return map[string]map[string]map[string]string{
		"pending": make(map[string]map[string]string),
		"queued":  make(map[string]map[string]string),
	}
}

func (s *TxPoolService) Status() map[string]hexutil.Uint {
	return map[string]hexutil.Uint{
		"pending": hexutil.Uint(0),
		"queued":  hexutil.Uint(0),
	}
}
