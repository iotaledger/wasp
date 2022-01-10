// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
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

func (w *WaspClientBackend) Signer() *cryptolib.KeyPair {
	return w.ChainClient.KeyPair
}

func (w *WaspClientBackend) PostOnLedgerRequest(scName, funName string, transfer *iscp.Assets, args dict.Dict) error {
	panic("TODO implement")
	// tx, err := w.ChainClient.Post1Request(iscp.Hn(scName), iscp.Hn(funName), chainclient.PostRequestParams{
	// 	Allowance: transfer,
	// 	Args:     args,
	// })
	// if err != nil {
	// 	return err
	// }
	// err = w.ChainClient.WaspClient.WaitUntilAllRequestsProcessed(w.ChainClient.ChainID, tx, 1*time.Minute)
	// if err != nil {
	// 	return err
	// }
	// for _, reqID := range request.RequestsInTransaction(w.ChainClient.ChainID, tx) {
	// 	return w.ChainClient.CheckRequestResult(reqID)
	// }
	// panic("should not reach here")
}

func (w *WaspClientBackend) PostOffLedgerRequest(scName, funName string, transfer *iscp.Assets, args dict.Dict) error {
	req, err := w.ChainClient.PostOffLedgerRequest(iscp.Hn(scName), iscp.Hn(funName), chainclient.PostRequestParams{
		Transfer: transfer,
		Args:     args,
	})
	if err != nil {
		return err
	}
	err = w.ChainClient.WaspClient.WaitUntilRequestProcessed(w.ChainClient.ChainID, req.ID(), 1*time.Minute)
	if err != nil {
		return err
	}
	return w.ChainClient.CheckRequestResult(req.ID())
}

func (w *WaspClientBackend) CallView(scName, funName string, args dict.Dict) (dict.Dict, error) {
	return w.ChainClient.CallView(iscp.Hn(scName), funName, args)
}
