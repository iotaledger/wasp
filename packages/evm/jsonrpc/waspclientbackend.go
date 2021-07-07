// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type WaspClientBackend struct {
	ChainClient *chainclient.Client
}

var _ ChainBackend = &WaspClientBackend{}

func NewWaspClientBackend(chainClient *chainclient.Client) *WaspClientBackend {
	return &WaspClientBackend{
		ChainClient: chainClient,
	}
}

func (w *WaspClientBackend) Signer() *ed25519.KeyPair {
	return w.ChainClient.KeyPair
}

func (w *WaspClientBackend) PostOnLedgerRequest(scName, funName string, transfer map[ledgerstate.Color]uint64, args dict.Dict) error {
	tx, err := w.ChainClient.Post1Request(coretypes.Hn(scName), coretypes.Hn(funName), chainclient.PostRequestParams{
		Transfer: ledgerstate.NewColoredBalances(transfer),
		Args:     requestargs.New(nil).AddEncodeSimpleMany(args),
	})
	if err != nil {
		return err
	}
	return w.ChainClient.WaspClient.WaitUntilAllRequestsProcessed(w.ChainClient.ChainID, tx, 1*time.Minute)
}

func (w *WaspClientBackend) PostOffLedgerRequest(scName, funName string, transfer map[ledgerstate.Color]uint64, args dict.Dict) error {
	req, err := w.ChainClient.PostOffLedgerRequest(coretypes.Hn(scName), coretypes.Hn(funName), chainclient.PostRequestParams{
		Transfer: ledgerstate.NewColoredBalances(transfer),
		Args:     requestargs.New().AddEncodeSimpleMany(args),
	})
	if err != nil {
		return err
	}
	return w.ChainClient.WaspClient.WaitUntilRequestProcessed(&w.ChainClient.ChainID, req.ID(), 1*time.Minute)
}

func (w *WaspClientBackend) CallView(scName, funName string, args dict.Dict) (dict.Dict, error) {
	return w.ChainClient.CallView(coretypes.Hn(scName), funName, args)
}
