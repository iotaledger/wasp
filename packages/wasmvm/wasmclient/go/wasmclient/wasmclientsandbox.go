// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"errors"
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

var (
	cvt          wasmhost.WasmConvertor
	HrpForClient = iotago.NetworkPrefix("")
)

func (s *WasmClientContext) FnCall(req *wasmrequests.CallRequest) []byte {
	s.eventReceived = false

	if req.Contract != s.scHname {
		s.Err = fmt.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}

	res, err := s.svcClient.CallViewByHname(s.chainID, req.Contract, req.Function, req.Params)
	if err != nil {
		s.Err = err
		return nil
	}
	return res
}

func (s *WasmClientContext) FnChainID() wasmtypes.ScChainID {
	return s.chainID
}

func (s *WasmClientContext) FnPost(req *wasmrequests.PostRequest) []byte {
	s.eventReceived = false

	if s.keyPair == nil {
		s.Err = errors.New("missing key pair")
		return nil
	}

	if req.ChainID != s.chainID {
		s.Err = fmt.Errorf("unknown chain id: %s", req.ChainID.String())
		return nil
	}

	if req.Contract != s.scHname {
		s.Err = fmt.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}

	scAssets := wasmlib.NewScAssets(req.Transfer)
	s.nonce++
	s.ReqID, s.Err = s.svcClient.PostRequest(req.ChainID, req.Contract, req.Function, req.Params, scAssets, s.keyPair, s.nonce)
	return nil
}

func ClientBech32Decode(bech32 string) wasmtypes.ScAddress {
	hrp, addr, err := iotago.ParseBech32(bech32)
	if err != nil {
		panic(err)
	}
	if hrp != HrpForClient {
		panic("invalid protocol prefix: " + string(hrp))
	}
	return cvt.ScAddress(addr)
}

func ClientBech32Encode(scAddress wasmtypes.ScAddress) string {
	addr := cvt.IscAddress(&scAddress)
	return addr.Bech32(HrpForClient)
}

func ClientHashName(name string) wasmtypes.ScHname {
	hName := isc.Hn(name)
	return cvt.ScHname(hName)
}
