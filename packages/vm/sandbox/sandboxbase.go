// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type Sandboxbase struct {
	Ctx       execution.WaspContext
	assertObj *assert.Assert
}

var _ iscp.SandboxBase = &Sandboxbase{}

func (s *Sandboxbase) assert() *assert.Assert {
	if s.assertObj == nil {
		s.assertObj = assert.NewAssert(s.Ctx)
	}
	return s.assertObj
}

func (s *Sandboxbase) AccountID() *iscp.AgentID {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.AccountID()
}

func (s *Sandboxbase) BalanceIotas() uint64 {
	s.Ctx.GasBurn(gas.BurnCodeGetBalance)
	return s.Ctx.GetIotaBalance(s.AccountID())
}

func (s *Sandboxbase) BalanceNativeToken(id *iotago.NativeTokenID) *big.Int {
	s.Ctx.GasBurn(gas.BurnCodeGetBalance)
	return s.Ctx.GetNativeTokenBalance(s.AccountID(), id)
}

func (s *Sandboxbase) Assets() *iscp.Assets {
	s.Ctx.GasBurn(gas.BurnCodeGetBalance)
	return s.Ctx.GetAssets(s.AccountID())
}

func (s *Sandboxbase) ChainID() *iscp.ChainID {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.ChainID()
}

func (s *Sandboxbase) ChainOwnerID() *iscp.AgentID {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.ChainOwnerID()
}

func (s *Sandboxbase) Contract() iscp.Hname {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.CurrentContractHname()
}

func (s *Sandboxbase) ContractCreator() *iscp.AgentID {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.ContractCreator()
}

func (s *Sandboxbase) Timestamp() int64 {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.Timestamp()
}

func (s *Sandboxbase) Log() iscp.LogInterface {
	// TODO should Log be disabled for wasm contracts? not much of a point in exposing internal logging
	return s.Ctx
}

func (s *Sandboxbase) Params() *iscp.Params {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.Params()
}

func (s *Sandboxbase) Utils() iscp.Utils {
	return NewUtils(s.Gas())
}

func (s *Sandboxbase) Gas() iscp.Gas {
	return s
}

func (s *Sandboxbase) Burn(burnCode gas.BurnCode, par ...uint64) {
	s.Ctx.GasBurn(burnCode, par...)
}

func (s *Sandboxbase) Budget() uint64 {
	return s.Ctx.GasBudgetLeft()
}

// -- helper methods
func (s *Sandboxbase) Requiref(cond bool, format string, args ...interface{}) {
	s.assert().Requiref(cond, format, args...)
}

func (s *Sandboxbase) RequireNoError(err error, str ...string) {
	s.assert().RequireNoError(err, str...)
}
