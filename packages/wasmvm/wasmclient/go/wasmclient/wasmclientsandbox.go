// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
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
	s.ReqID, s.Err = s.svcClient.PostRequest(req.ChainID, req.Contract, req.Function, req.Params, scAssets, s.keyPair)
	return nil
}
