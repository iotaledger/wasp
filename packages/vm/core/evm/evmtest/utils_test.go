// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

func (i *iscTestContractInstance) getChainID() isc.ChainID {
	var v isc.ChainID
	require.NoError(i.chain.t, i.callView("getChainID", nil, &v))
	return v
}

func (s *storageContractInstance) retrieve() uint32 {
	var v uint32
	require.NoError(s.chain.t, s.callView("retrieve", nil, &v))
	return v
}

func (s *storageContractInstance) store(n uint32, opts ...ethCallOptions) (res CallFnResult, err error) {
	return s.CallFn(opts, "store", n)
}

func (e *erc20ContractInstance) balanceOf(addr common.Address) *big.Int {
	v := new(big.Int)
	require.NoError(e.chain.t, e.callView("balanceOf", []interface{}{addr}, &v))
	return v
}

func (e *erc20ContractInstance) totalSupply() *big.Int {
	v := new(big.Int)
	require.NoError(e.chain.t, e.callView("totalSupply", nil, &v))
	return v
}

func (e *erc20ContractInstance) transfer(recipientAddress common.Address, amount *big.Int, opts ...ethCallOptions) (res CallFnResult, err error) {
	return e.CallFn(opts, "transfer", recipientAddress, amount)
}

func (l *loopContractInstance) loop(opts ...ethCallOptions) (res CallFnResult, err error) {
	return l.CallFn(opts, "loop")
}

func (f *fibonacciContractInstance) fib(n uint32, opts ...ethCallOptions) (res CallFnResult, err error) {
	return f.CallFn(opts, "fib", n)
}

func generateEthereumKey(t testing.TB) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}
