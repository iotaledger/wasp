// package jsonrpc implements JSON-RPC endpoints according to
// https://eth.wiki/json-rpc/API
package jsonrpc

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/evm"
	"golang.org/x/crypto/sha3"
	"golang.org/x/xerrors"
)

type EthService struct {
	evmChain     *EVMChain
	tangleSigner *ed25519.KeyPair
	accounts     *AccountManager
}

func NewEthService(evmChain *EVMChain, tangleSigner *ed25519.KeyPair, accounts *AccountManager) *EthService {
	return &EthService{evmChain, tangleSigner, accounts}
}

func (e *EthService) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(eth.ETH65)
}

func (e *EthService) GetTransactionCount(address common.Address, blockNumber rpc.BlockNumber) (hexutil.Uint64, error) {
	n, err := e.evmChain.TransactionCount(address, parseBlockNumber(blockNumber))
	if err != nil {
		return 0, err
	}
	return hexutil.Uint64(n), nil
}

func (e *EthService) BlockNumber() (*hexutil.Big, error) {
	n, err := e.evmChain.BlockNumber()
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(n), nil
}

func (e *EthService) GetBlockByNumber(blockNumber rpc.BlockNumber, full bool) (map[string]interface{}, error) {
	block, err := e.evmChain.BlockByNumber(parseBlockNumber(blockNumber))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return RPCMarshalBlock(block, true, full)
}

func (e *EthService) GetBlockByHash(hash common.Hash, full bool) (map[string]interface{}, error) {
	block, err := e.evmChain.BlockByHash(hash)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return RPCMarshalBlock(block, true, full)
}

func (e *EthService) GetTransactionByHash(hash common.Hash) (*RPCTransaction, error) {
	tx, blockHash, blockNumber, index, err := e.evmChain.TransactionByHash(hash)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, nil
	}
	return newRPCTransaction(tx, blockHash, blockNumber, index), err
}

func (e *EthService) GetTransactionByBlockHashAndIndex(blockHash common.Hash, index hexutil.Uint) (*RPCTransaction, error) {
	tx, _, blockNumber, _, err := e.evmChain.TransactionByBlockHashAndIndex(blockHash, uint64(index))
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, nil
	}
	return newRPCTransaction(tx, blockHash, blockNumber, uint64(index)), err
}

func (e *EthService) GetTransactionByBlockNumberAndIndex(blockNumberOrTag rpc.BlockNumber, index hexutil.Uint) (*RPCTransaction, error) {
	tx, blockHash, blockNumber, _, err := e.evmChain.TransactionByBlockNumberAndIndex(parseBlockNumber(blockNumberOrTag), uint64(index))
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, nil
	}
	return newRPCTransaction(tx, blockHash, blockNumber, uint64(index)), err
}

func (e *EthService) GetBalance(address common.Address, blockNumber rpc.BlockNumber) (*hexutil.Big, error) {
	bal, err := e.evmChain.Balance(address, parseBlockNumber(blockNumber))
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(bal), nil
}

func (e *EthService) GetCode(address common.Address, blockNumber rpc.BlockNumber) (hexutil.Bytes, error) {
	code, err := e.evmChain.Code(address, parseBlockNumber(blockNumber))
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(code), nil
}

func (e *EthService) GetTransactionReceipt(txHash common.Hash) (map[string]interface{}, error) {
	r, err := e.evmChain.TransactionReceipt(txHash)
	if err != nil {
		return nil, err
	}
	return RPCMarshalReceipt(r), nil
}

func (e *EthService) SendRawTransaction(txBytes hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(txBytes, tx); err != nil {
		return common.Hash{}, err
	}
	if err := e.evmChain.SendTransaction(e.tangleSigner, tx); err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}

func (e *EthService) Call(args *RPCCallArgs, blockNumber rpc.BlockNumber) (hexutil.Bytes, error) {
	ret, err := e.evmChain.CallContract(args.parse(), parseBlockNumber(blockNumber))
	return hexutil.Bytes(ret), err
}

func (e *EthService) EstimateGas(args *RPCCallArgs) (hexutil.Uint64, error) {
	gas, err := e.evmChain.EstimateGas(args.parse())
	return hexutil.Uint64(gas), err
}

func (e *EthService) GetStorageAt(address common.Address, key common.Hash, blockNumber rpc.BlockNumber) (hexutil.Bytes, error) {
	ret, err := e.evmChain.StorageAt(address, key, parseBlockNumber(blockNumber))
	return hexutil.Bytes(ret), err
}

func (e *EthService) GetBlockTransactionCountByHash(blockHash common.Hash) (hexutil.Uint, error) {
	ret, err := e.evmChain.BlockTransactionCountByHash(blockHash)
	return hexutil.Uint(ret), err
}

func (e *EthService) GetBlockTransactionCountByNumber(blockNumber rpc.BlockNumber) (hexutil.Uint, error) {
	ret, err := e.evmChain.BlockTransactionCountByNumber(parseBlockNumber(blockNumber))
	return hexutil.Uint(ret), err
}

func (e *EthService) GetUncleCountByBlockHash(blockHash common.Hash) hexutil.Uint {
	return hexutil.Uint(0) // no uncles are ever generated in evmchain contract
}

func (e *EthService) GetUncleCountByBlockNumber(blockNumber rpc.BlockNumber) hexutil.Uint {
	return hexutil.Uint(0) // no uncles are ever generated in evmchain contract
}

func (e *EthService) GetUncleByBlockHashAndIndex(blockHash common.Hash, index hexutil.Uint) map[string]interface{} {
	return nil // no uncles are ever generated in evmchain contract
}

func (e *EthService) GetUncleByBlockNumberAndIndex(blockNumberOrTag rpc.BlockNumber, index hexutil.Uint) map[string]interface{} {
	return nil // no uncles are ever generated in evmchain contract
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
	if err := e.evmChain.SendTransaction(e.tangleSigner, tx); err != nil {
		return common.Hash{}, err
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
	return types.SignTx(args.toTransaction(), evm.Signer(), account)
}

func (e *EthService) GetLogs(q *RPCFilterQuery) ([]*types.Log, error) {
	return e.evmChain.Logs((*ethereum.FilterQuery)(q))
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

type NetService struct{}

func NewNetService() *NetService {
	return &NetService{}
}

func (s *NetService) Version() string {
	return "1074" // IOTA -- get it?
}

func (s *NetService) Listening() bool         { return true }
func (s *NetService) PeerCount() hexutil.Uint { return 0 }

type Web3Service struct{}

func NewWeb3Service() *Web3Service {
	return &Web3Service{}
}

func (s *Web3Service) ClientVersion() string {
	return "evmproxy"
}

func (s *Web3Service) Sha3(input hexutil.Bytes) hexutil.Bytes {
	return crypto.Keccak256(input)
}
