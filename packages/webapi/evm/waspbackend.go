// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/chainutil"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

type jsonRPCWaspBackend struct {
	chain      chain.Chain
	nodePubKey *cryptolib.PublicKey
	requestIDs sync.Map
	baseToken  *parameters.BaseToken
}

var _ jsonrpc.ChainBackend = &jsonRPCWaspBackend{}

func newWaspBackend(ch chain.Chain, nodePubKey *cryptolib.PublicKey, baseToken *parameters.BaseToken) *jsonRPCWaspBackend {
	return &jsonRPCWaspBackend{
		chain:      ch,
		nodePubKey: nodePubKey,
		baseToken:  baseToken,
	}
}

func (b *jsonRPCWaspBackend) RequestIDByTransactionHash(txHash common.Hash) (isc.RequestID, bool) {
	// TODO: should this be stored in the chain state instead of a volatile cache?
	r, ok := b.requestIDs.Load(txHash)
	if !ok {
		return isc.RequestID{}, false
	}
	return r.(isc.RequestID), true
}

func (b *jsonRPCWaspBackend) EVMGasRatio() (util.Ratio32, error) {
	// TODO: Cache the gas ratio?
	ret, err := b.ISCCallView(evm.Contract.Name, evm.FuncGetGasRatio.Name, nil)
	if err != nil {
		return util.Ratio32{}, err
	}
	return codec.DecodeRatio32(ret.MustGet(evm.FieldResult))
}

func (b *jsonRPCWaspBackend) EVMSendTransaction(tx *types.Transaction) error {
	// Ensure the transaction has more gas than the basic Ethereum tx fee.
	intrinsicGas, err := core.IntrinsicGas(tx.Data(), tx.AccessList(), tx.To() == nil, true, true)
	if err != nil {
		return err
	}
	if tx.Gas() < intrinsicGas {
		return core.ErrIntrinsicGas
	}

	req, err := isc.NewEVMOffLedgerRequest(b.chain.ID(), tx)
	if err != nil {
		return err
	}
	b.chain.EnqueueOffLedgerRequestMsg(&messages.OffLedgerRequestMsgIn{
		OffLedgerRequestMsg: messages.OffLedgerRequestMsg{
			ChainID: b.chain.ID(),
			Req:     req,
		},
		SenderPubKey: b.nodePubKey,
	})

	// store the request ID so that the user can query it later (if the
	// Etheeum tx fails, the Ethereum receipt is never generated).
	txHash := tx.Hash()
	b.requestIDs.Store(txHash, req.ID())
	go b.evictWhenExpired(txHash)

	return nil
}

func (b *jsonRPCWaspBackend) evictWhenExpired(txHash common.Hash) {
	time.Sleep(1 * time.Hour)
	b.requestIDs.Delete(txHash)
}

func (b *jsonRPCWaspBackend) EVMEstimateGas(callMsg ethereum.CallMsg) (uint64, error) {
	return chainutil.SimulateCall(b.chain, callMsg)
}

func (b *jsonRPCWaspBackend) ISCCallView(scName, funName string, args dict.Dict) (dict.Dict, error) {
	return chainutil.CallView(b.chain, isc.Hn(scName), isc.Hn(funName), args)
}

func (b *jsonRPCWaspBackend) BaseToken() *parameters.BaseToken {
	return b.baseToken
}
