// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/dividend/go/dividend"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func dividendMember(ctx *wasmsolo.SoloContext, agent *wasmsolo.SoloAgent, factor uint64) {
	member := dividend.ScFuncs.Member(ctx)
	member.Params.Address().SetValue(agent.ScAddress())
	member.Params.Factor().SetValue(factor)
	member.Func.TransferIotas(1111).Post()
}

func dividendDivide(ctx *wasmsolo.SoloContext, amount uint64) {
	divide := dividend.ScFuncs.Divide(ctx)
	divide.Func.TransferIotas(amount).Post()
}

func dividendGetFactor(ctx *wasmsolo.SoloContext, member *wasmsolo.SoloAgent) uint64 {
	getFactor := dividend.ScFuncs.GetFactor(ctx)
	getFactor.Params.Address().SetValue(member.ScAddress())
	getFactor.Func.Call()
	value := getFactor.Results.Factor().Value()
	return value
}

func TestDeploy(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)
	require.NoError(t, ctx.ContractExists(dividend.ScName))
}

func TestAddMemberOk(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)
	ctx.Accounts()

	member1 := ctx.NewSoloAgent()
	ctx.Accounts(member1)

	dividendMember(ctx, member1, 100)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1)
}

func TestAddMemberFailMissingAddress(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	member := dividend.ScFuncs.Member(ctx)
	member.Params.Factor().SetValue(100)
	member.Func.TransferIotas(1).Post()
	require.Error(t, ctx.Err)
	require.Contains(t, ctx.Err.Error(), "missing mandatory address")
}

func TestAddMemberFailMissingFactor(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	member1 := ctx.NewSoloAgent()
	member := dividend.ScFuncs.Member(ctx)
	member.Params.Address().SetValue(member1.ScAddress())
	member.Func.TransferIotas(1).Post()
	require.Error(t, ctx.Err)
	require.Contains(t, ctx.Err.Error(), "missing mandatory factor")
}

func TestDivide1Member(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	accountBalance := ctx.Balance(ctx.Account())
	chainBalance := ctx.Balance(ctx.ChainAccount())
	originatorBalance := ctx.Balance(ctx.Originator())

	member1 := ctx.NewSoloAgent()
	ctx.Accounts(member1)
	member1Balance := ctx.Balance(member1)

	dividendMember(ctx, member1, 1000)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1)

	chainBalance += ctx.GasFee
	originatorBalance += 1111 - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, member1Balance, ctx.Balance(member1))

	dividendDivide(ctx, 999)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1)

	chainBalance += ctx.GasFee
	originatorBalance -= ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, member1Balance+999, ctx.Balance(member1))
}

func TestDivide2Members(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	accountBalance := ctx.Balance(ctx.Account())
	chainBalance := ctx.Balance(ctx.ChainAccount())
	originatorBalance := ctx.Balance(ctx.Originator())

	member1 := ctx.NewSoloAgent()
	ctx.Accounts(member1)

	dividendMember(ctx, member1, 250)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1)

	chainBalance += ctx.GasFee
	originatorBalance += 1111 - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, 0, ctx.Balance(member1))

	member2 := ctx.NewSoloAgent()
	ctx.Accounts(member1, member2)

	dividendMember(ctx, member2, 750)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1, member2)

	chainBalance += ctx.GasFee
	originatorBalance += 1111 - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, 0, ctx.Balance(member1))
	require.EqualValues(t, 0, ctx.Balance(member2))

	dividendDivide(ctx, 999)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1, member2)

	remain := uint64(999) - 999*250/1000 - 999*750/1000
	chainBalance += ctx.GasFee
	originatorBalance += remain - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, 999*250/1000, ctx.Balance(member1))
	require.EqualValues(t, 999*750/1000, ctx.Balance(member2))
}

func TestDivide3Members(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	accountBalance := ctx.Balance(ctx.Account())
	chainBalance := ctx.Balance(ctx.ChainAccount())
	originatorBalance := ctx.Balance(ctx.Originator())

	member1 := ctx.NewSoloAgent()
	ctx.Accounts(member1)

	dividendMember(ctx, member1, 250)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1)

	chainBalance += ctx.GasFee
	originatorBalance += 1111 - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, 0, ctx.Balance(member1))

	member2 := ctx.NewSoloAgent()
	ctx.Accounts(member1, member2)

	dividendMember(ctx, member2, 500)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1, member2)

	chainBalance += ctx.GasFee
	originatorBalance += 1111 - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, 0, ctx.Balance(member1))
	require.EqualValues(t, 0, ctx.Balance(member2))

	member3 := ctx.NewSoloAgent()
	ctx.Accounts(member1, member2, member3)

	dividendMember(ctx, member3, 750)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1, member2, member3)

	chainBalance += ctx.GasFee
	originatorBalance += 1111 - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, 0, ctx.Balance(member1))
	require.EqualValues(t, 0, ctx.Balance(member2))
	require.EqualValues(t, 0, ctx.Balance(member3))

	dividendDivide(ctx, 999)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1, member2, member3)

	remain := uint64(999) - 999*250/1500 - 999*500/1500 - 999*750/1500
	chainBalance += ctx.GasFee
	originatorBalance += remain - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, 999*250/1500, ctx.Balance(member1))
	require.EqualValues(t, 999*500/1500, ctx.Balance(member2))
	require.EqualValues(t, 999*750/1500, ctx.Balance(member3))

	dividendDivide(ctx, 1234)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1, member2, member3)

	remain = uint64(1234) - 1234*250/1500 - 1234*500/1500 - 1234*750/1500
	chainBalance += ctx.GasFee
	originatorBalance += remain - ctx.GasFee
	require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
	require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
	require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	require.EqualValues(t, 999*250/1500+1234*250/1500, ctx.Balance(member1))
	require.EqualValues(t, 999*500/1500+1234*500/1500, ctx.Balance(member2))
	require.EqualValues(t, 999*750/1500+1234*750/1500, ctx.Balance(member3))
}

func TestGetFactor(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	member1 := ctx.NewSoloAgent()
	dividendMember(ctx, member1, 25)
	require.NoError(t, ctx.Err)

	member2 := ctx.NewSoloAgent()
	dividendMember(ctx, member2, 50)
	require.NoError(t, ctx.Err)

	member3 := ctx.NewSoloAgent()
	dividendMember(ctx, member3, 75)
	require.NoError(t, ctx.Err)
	ctx.Accounts(member1, member2, member3)

	value := dividendGetFactor(ctx, member3)
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 75, value)

	value = dividendGetFactor(ctx, member2)
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 50, value)

	value = dividendGetFactor(ctx, member1)
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 25, value)
}
