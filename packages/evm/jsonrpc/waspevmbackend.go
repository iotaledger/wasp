// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type WaspEVMBackend struct {
	chain      chain.Chain
	nodePubKey *cryptolib.PublicKey
	requestIDs sync.Map
	baseToken  *parameters.BaseToken
}

var _ ChainBackend = &WaspEVMBackend{}

func NewWaspEVMBackend(ch chain.Chain, nodePubKey *cryptolib.PublicKey, baseToken *parameters.BaseToken) *WaspEVMBackend {
	return &WaspEVMBackend{
		chain:      ch,
		nodePubKey: nodePubKey,
		baseToken:  baseToken,
	}
}

func (b *WaspEVMBackend) RequestIDByTransactionHash(txHash common.Hash) (isc.RequestID, bool) {
	// TODO: should this be stored in the chain state instead of a volatile cache?
	r, ok := b.requestIDs.Load(txHash)
	if !ok {
		return isc.RequestID{}, false
	}
	return r.(isc.RequestID), true
}

func (b *WaspEVMBackend) EVMGasRatio() (util.Ratio32, error) {
	// TODO: Cache the gas ratio?
	ret, err := b.ISCCallView(b.ISCLatestState(), governance.Contract.Name, governance.ViewGetEVMGasRatio.Name, nil)
	if err != nil {
		return util.Ratio32{}, err
	}
	return codec.DecodeRatio32(ret.MustGet(governance.ParamEVMGasRatio))
}

func (b *WaspEVMBackend) EVMSendTransaction(tx *types.Transaction) error {
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
	b.chain.ReceiveOffLedgerRequest(req, b.nodePubKey)

	// store the request ID so that the user can query it later (if the
	// Etheeum tx fails, the Ethereum receipt is never generated).
	txHash := tx.Hash()
	b.requestIDs.Store(txHash, req.ID())
	go b.evictWhenExpired(txHash)

	return nil
}

func (b *WaspEVMBackend) evictWhenExpired(txHash common.Hash) {
	time.Sleep(1 * time.Hour)
	b.requestIDs.Delete(txHash)
}

func (b *WaspEVMBackend) EVMEstimateGas(callMsg ethereum.CallMsg) (uint64, error) {
	return chainutil.EstimateGas(b.chain, callMsg)
}

func (b *WaspEVMBackend) EVMGasPrice() *big.Int {
	latestState := b.ISCLatestState()
	res, err := chainutil.CallView(latestState, b.chain, governance.Contract.Hname(), governance.ViewGetFeePolicy.Hname(), nil)
	if err != nil {
		panic(fmt.Sprintf("couldn't call gasFeePolicy view: %s ", err.Error()))
	}
	feePolicy, err := gas.FeePolicyFromBytes(res.MustGet(governance.ParamFeePolicyBytes))
	if err != nil {
		panic(fmt.Sprintf("couldn't decode fee policy: %s ", err.Error()))
	}
	res, err = chainutil.CallView(latestState, b.chain, governance.Contract.Hname(), governance.ViewGetEVMGasRatio.Hname(), nil)
	if err != nil {
		panic(fmt.Sprintf("couldn't call getGasRatio view: %s ", err.Error()))
	}
	gasRatio := codec.MustDecodeRatio32(res.MustGet(governance.ParamEVMGasRatio))

	// convert to wei (18 decimals)
	decimalsDifference := 18 - parameters.L1().BaseToken.Decimals
	price := big.NewInt(10)
	price.Exp(price, new(big.Int).SetUint64(uint64(decimalsDifference)), nil)

	price.Mul(price, new(big.Int).SetUint64(uint64(gasRatio.A)))
	price.Div(price, new(big.Int).SetUint64(uint64(gasRatio.B)))
	price.Div(price, new(big.Int).SetUint64(feePolicy.GasPerToken))

	return price
}

func (b *WaspEVMBackend) ISCCallView(chainState state.State, scName, funName string, args dict.Dict) (dict.Dict, error) {
	return chainutil.CallView(chainState, b.chain, isc.Hn(scName), isc.Hn(funName), args)
}

func (b *WaspEVMBackend) BaseToken() *parameters.BaseToken {
	return b.baseToken
}

// ISCLatestState implements jsonrpc.ChainBackend
func (b *WaspEVMBackend) ISCLatestState() state.State {
	latestState, err := b.chain.LatestState(chain.LatestState)
	if err != nil {
		panic(fmt.Sprintf("couldn't get latest block index: %s ", err.Error()))
	}
	return latestState
}

// ISCLatestState implements jsonrpc.ChainBackend
func (b *WaspEVMBackend) ISCStateByBlockIndex(blockIndex uint32) (state.State, error) {
	latestState, err := b.chain.LatestState(chain.LatestState)
	if err != nil {
		return nil, fmt.Errorf("couldn't get latest state: %s", err.Error())
	}
	if latestState.BlockIndex() == blockIndex {
		return latestState, nil
	}
	return b.chain.Store().StateByIndex(blockIndex)
}
