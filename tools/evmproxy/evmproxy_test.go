package main

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/tools/evmproxy/service"
	"github.com/stretchr/testify/require"
)

type env struct {
	t      *testing.T
	server *rpc.Server
	client *ethclient.Client
}

func newEnv(t *testing.T) *env {
	soloEVMChain := service.NewEVMChain(service.NewSoloBackend(core.GenesisAlloc{
		faucetAddress: {Balance: faucetSupply},
	}))

	rpcsrv := NewRPCServer(soloEVMChain)
	t.Cleanup(rpcsrv.Stop)

	client := ethclient.NewClient(rpc.DialInProc(rpcsrv))
	t.Cleanup(client.Close)

	return &env{t, rpcsrv, client}
}

var requestFundsAmount = big.NewInt(1e18) // 1 ETH

func (e *env) requestFunds(target common.Address) {
	nonce, err := e.client.NonceAt(context.Background(), faucetAddress, nil)
	require.NoError(e.t, err)
	tx, err := types.SignTx(
		types.NewTransaction(nonce, target, requestFundsAmount, evm.GasLimit, evm.GasPrice, nil),
		evm.Signer(),
		faucetKey,
	)
	require.NoError(e.t, err)
	err = e.client.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)
}

func (e *env) nonceAt(address common.Address) uint64 {
	nonce, err := e.client.NonceAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return nonce
}

func (e *env) blockNumber() uint64 {
	blockNumber, err := e.client.BlockNumber(context.Background())
	require.NoError(e.t, err)
	return blockNumber
}

func (e *env) balance(address common.Address) *big.Int {
	bal, err := e.client.BalanceAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return bal
}

func TestRPCRequestFunds(t *testing.T) {
	env := newEnv(t)

	receiver, err := crypto.GenerateKey()
	require.NoError(t, err)
	receiverAddress := crypto.PubkeyToAddress(receiver.PublicKey)

	require.EqualValues(t, 0, env.blockNumber())
	require.EqualValues(t, 0, env.nonceAt(faucetAddress))
	require.Zero(t, big.NewInt(0).Cmp(env.balance(receiverAddress)))

	env.requestFunds(receiverAddress)

	require.EqualValues(t, 1, env.blockNumber())
	require.EqualValues(t, 1, env.nonceAt(faucetAddress))
	require.Zero(t, big.NewInt(1e18).Cmp(env.balance(receiverAddress)))
}
