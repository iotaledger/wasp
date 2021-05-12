// package service implements JSON-RPC endpoints according to
// https://eth.wiki/json-rpc/API
package service

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type EVMChain interface {
	BlockNumber() *big.Int
	Balance(address common.Address, blockNumber *big.Int) *big.Int
	BlockByNumber(blockNumber *big.Int) *types.Block
}

type EthService struct {
	chain EVMChain
}

func NewEthService(chain EVMChain) *EthService {
	return &EthService{chain}
}

func (e *EthService) BlockNumber() *hexutil.Big {
	return (*hexutil.Big)(e.chain.BlockNumber())
}

func (e *EthService) GetBlockByNumber(blockNumber rpc.BlockNumber, full bool) (map[string]interface{}, error) {
	block := e.chain.BlockByNumber(parseBlockNumber(blockNumber))
	if block == nil {
		return nil, nil
	}
	return RPCMarshalBlock(block, true, full)
}

func (e *EthService) GetBalance(address common.Address, blockNumber rpc.BlockNumber) *hexutil.Big {
	return (*hexutil.Big)(e.chain.Balance(address, parseBlockNumber(blockNumber)))
}

type NetService struct {
	chain EVMChain
}

func NewNetService(chain EVMChain) *NetService {
	return &NetService{chain}
}

func (e *NetService) Version() string {
	return "1074" // IOTA -- get it?
}
