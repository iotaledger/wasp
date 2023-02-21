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
	if req.Contract != s.scHname {
		s.Err = fmt.Errorf("unknown contract: %s", req.Contract.String())
		return nil
	}

	res, err := s.svcClient.CallViewByHname(req.Contract, req.Function, req.Params)
	if err != nil {
		s.Err = err
		return nil
	}
	return res
}

func (s *WasmClientContext) FnChainID() wasmtypes.ScChainID {
	return s.CurrentChainID()
}

func (s *WasmClientContext) FnPost(req *wasmrequests.PostRequest) []byte {
	if s.keyPair == nil {
		s.Err = errors.New("missing key pair")
		return nil
	}

	if req.ChainID != s.CurrentChainID() {
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

func clientBech32Decode(bech32 string) wasmtypes.ScAddress {
	hrp, addr, err := iotago.ParseBech32(bech32)
	if err != nil {
		panic(err)
	}
	if hrp != HrpForClient {
		panic("invalid protocol prefix: " + string(hrp))
	}
	return cvt.ScAddress(addr)
}

func clientBech32Encode(scAddress wasmtypes.ScAddress) string {
	addr := cvt.IscAddress(&scAddress)
	return addr.Bech32(HrpForClient)
}

func clientHashName(name string) wasmtypes.ScHname {
	hName := isc.Hn(name)
	return cvt.ScHname(hName)
}

func SetSandboxWrappers(chainID string) error {
	if HrpForClient != "" {
		return nil
	}

	// local client implementations for some sandbox functions
	wasmtypes.Bech32Decode = clientBech32Decode
	wasmtypes.Bech32Encode = clientBech32Encode
	wasmtypes.HashName = clientHashName

	// set the network prefix for the current network
	hrp, _, err := iotago.ParseBech32(chainID)
	if err != nil {
		return err
	}
	if HrpForClient != hrp && HrpForClient != "" {
		panic("WasmClient can only connect to one Tangle network per app")
	}
	HrpForClient = hrp
	return nil
}
