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

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

func (i *iscTestContractInstance) getChainID() isc.ChainID {
	var v iscmagic.ISCChainID
	require.NoError(i.chain.t, i.callView("getChainID", nil, &v))
	return v.MustUnwrap()
}

func (i *iscTestContractInstance) triggerEvent(s string) (res callFnResult, err error) {
	return i.callFn(nil, "triggerEvent", s)
}

func (i *iscTestContractInstance) triggerEventFail(s string, opts ...ethCallOptions) (res callFnResult, err error) {
	return i.callFn(opts, "triggerEventFail", s)
}

func (s *storageContractInstance) retrieve() uint32 {
	var v uint32
	require.NoError(s.chain.t, s.callView("retrieve", nil, &v))
	return v
}

func (s *storageContractInstance) store(n uint32, opts ...ethCallOptions) (res callFnResult, err error) {
	return s.callFn(opts, "store", n)
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

func (e *erc20ContractInstance) transfer(recipientAddress common.Address, amount *big.Int, opts ...ethCallOptions) (res callFnResult, err error) {
	return e.callFn(opts, "transfer", recipientAddress, amount)
}

func (l *loopContractInstance) loop(opts ...ethCallOptions) (res callFnResult, err error) {
	return l.callFn(opts, "loop")
}

func (f *fibonacciContractInstance) fib(n uint32, opts ...ethCallOptions) (res callFnResult, err error) {
	return f.callFn(opts, "fib", n)
}

func generateEthereumKey(t testing.TB) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}
