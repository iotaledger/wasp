// package service implements JSON-RPC endpoints according to
// https://eth.wiki/json-rpc/API
package service

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/evm"
)

type EVMChain interface {
	BlockNumber() *big.Int
	SendTransaction(tx *types.Transaction)
	Balance(address common.Address, blockNumber *big.Int) *big.Int
	BlockByNumber(blockNumber *big.Int) *types.Block
	TransactionCount(address common.Address, blockNumber *big.Int) uint64
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

type TestService struct {
	chain     EVMChain
	faucetKey *ecdsa.PrivateKey
}

func NewTestService(chain EVMChain, faucetKey *ecdsa.PrivateKey) *TestService {
	return &TestService{chain, faucetKey}
}

func (e *TestService) RequestFunds(address common.Address) error {
	nonce := e.chain.TransactionCount(crypto.PubkeyToAddress(e.faucetKey.PublicKey), nil)
	amount := big.NewInt(1e18) // 1 ETH
	tx, err := types.SignTx(
		types.NewTransaction(nonce, address, amount, evm.GasLimit, evm.GasPrice, nil),
		evm.Signer(),
		e.faucetKey,
	)
	if err != nil {
		return err
	}
	e.chain.SendTransaction(tx)
	return nil
}
