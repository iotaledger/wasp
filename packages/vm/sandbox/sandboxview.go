// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

type sandboxView struct {
	vmctx           *vmcontext.VMContext
	assertObj       *assert.Assert
	paramDecoderObj iscp.KVDecoder
}

func (s *sandboxView) assert() *assert.Assert {
	if s.assertObj == nil {
		s.assertObj = assert.NewAssert(s.vmctx)
	}
	return s.assertObj
}

func (s *sandboxView) paramDecoder() iscp.KVDecoder {
	if s.paramDecoderObj == nil {
		s.paramDecoderObj = kvdecoder.New(s.vmctx.Params(), s.Log())
	}
	return s.paramDecoderObj
}

func (s *sandboxView) Assets() *iscp.Assets {
	panic("implement me")
}

func (s *sandboxView) Timestamp() int64 {
	panic("implement me")
}

var _ iscp.SandboxView = &sandboxView{}

func init() {
	vmcontext.NewSandboxView = func(vmctx *vmcontext.VMContext) iscp.SandboxView {
		return &sandboxView{
			vmctx: vmctx,
		}
	}
}

func (s *sandboxView) AccountID() *iscp.AgentID {
	return s.vmctx.AccountID()
}

func (s *sandboxView) BalanceIotas() uint64 {
	panic("implement me")
}

func (s *sandboxView) BalanceNativeToken(id *iotago.NativeTokenID) *big.Int {
	panic("implement me")
}

func (s *sandboxView) Call(contractHname, entryPoint iscp.Hname, params dict.Dict) dict.Dict {
	return s.vmctx.Call(contractHname, entryPoint, params, nil)
}

func (s *sandboxView) ChainID() *iscp.ChainID {
	return s.vmctx.ChainID()
}

func (s *sandboxView) ChainOwnerID() *iscp.AgentID {
	return s.vmctx.ChainOwnerID()
}

func (s *sandboxView) Contract() iscp.Hname {
	return s.vmctx.CurrentContractHname()
}

func (s *sandboxView) ContractCreator() *iscp.AgentID {
	return s.vmctx.ContractCreator()
}

func (s *sandboxView) GetTimestamp() int64 {
	return s.vmctx.Timestamp()
}

func (s *sandboxView) Log() iscp.LogInterface {
	return s.vmctx
}

func (s *sandboxView) Params() dict.Dict {
	return s.vmctx.Params()
}

func (s *sandboxView) ParamDecoder() iscp.KVDecoder {
	return s.paramDecoder()
}

func (s *sandboxView) State() kv.KVStoreReader {
	return s.vmctx.State()
}

func (s *sandboxView) Utils() iscp.Utils {
	return NewUtils(s.Gas())
}

func (s *sandboxView) Gas() iscp.Gas {
	return s
}

func (s *sandboxView) Burn(burnCode gas.BurnCode, par ...uint64) {
	s.vmctx.GasBurn(burnCode, par...)
}

func (s *sandboxView) Budget() uint64 {
	return s.vmctx.GasBudgetLeft()
}

// helper methods

func (s *sandboxView) Requiref(cond bool, format string, args ...interface{}) {
	s.assert().Requiref(cond, format, args...)
}

func (s *sandboxView) RequireNoError(err error, str ...string) {
	s.assert().RequireNoError(err, str...)
}
