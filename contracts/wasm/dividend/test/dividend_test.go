// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/dividend/go/dividend"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func dividendMember(ctx *wasmsolo.SoloContext, agent *wasmsolo.SoloAgent, factor uint64) {
	member := dividend.ScFuncs.Member(ctx)
	member.Params.Address().SetValue(agent.ScAgentID().Address())
	member.Params.Factor().SetValue(factor)
	member.Func.Post()
}

func dividendDivide(ctx *wasmsolo.SoloContext, amount uint64) {
	divide := dividend.ScFuncs.Divide(ctx)
	divide.Func.TransferIotas(amount).Post()
}

func dividendGetFactor(ctx *wasmsolo.SoloContext, member *wasmsolo.SoloAgent) uint64 {
	getFactor := dividend.ScFuncs.GetFactor(ctx)
	getFactor.Params.Address().SetValue(member.ScAgentID().Address())
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

	member1 := ctx.NewSoloAgent()
	dividendMember(ctx, member1, 100)
	require.NoError(t, ctx.Err)
}

func TestAddMemberFailMissingAddress(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	member := dividend.ScFuncs.Member(ctx)
	member.Params.Factor().SetValue(100)
	member.Func.Post()
	require.Error(t, ctx.Err)
	require.Contains(t, ctx.Err.Error(), "missing mandatory address")
}

func TestAddMemberFailMissingFactor(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	member1 := ctx.NewSoloAgent()
	member := dividend.ScFuncs.Member(ctx)
	member.Params.Address().SetValue(member1.ScAgentID().Address())
	member.Func.Post()
	require.Error(t, ctx.Err)
	require.Contains(t, ctx.Err.Error(), "missing mandatory factor")
}

func TestDivide1Member(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	member1 := ctx.NewSoloAgent()
	bal := ctx.Balances(member1)

	dividendMember(ctx, member1, 1000)
	require.NoError(t, ctx.Err)

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.Dust - ctx.GasFee
	bal.VerifyBalances(t)

	toDivide := 1*iscp.Mi + 1
	dividendDivide(ctx, toDivide)
	require.NoError(t, ctx.Err)

	bal.Chain += ctx.GasFee
	bal.Originator -= ctx.GasFee
	bal.Add(member1, toDivide)
	bal.VerifyBalances(t)
}

func TestDivide2Members(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	member1 := ctx.NewSoloAgent()
	bal := ctx.Balances(member1)

	dividendMember(ctx, member1, 250)
	require.NoError(t, ctx.Err)

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.Dust - ctx.GasFee
	bal.VerifyBalances(t)

	member2 := ctx.NewSoloAgent()
	bal = ctx.Balances(member1, member2)

	dividendMember(ctx, member2, 750)
	require.NoError(t, ctx.Err)

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.Dust - ctx.GasFee
	bal.VerifyBalances(t)

	toDivide := 5*iscp.Mi + 1
	dividendDivide(ctx, toDivide)
	require.NoError(t, ctx.Err)

	remain := uint64(toDivide) - toDivide*250/1000 - toDivide*750/1000
	bal.Chain += ctx.GasFee
	bal.Originator += remain - ctx.GasFee
	bal.Add(member1, toDivide*250/1000)
	bal.Add(member2, toDivide*750/1000)
	bal.VerifyBalances(t)
}

func TestDivide3Members(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, dividend.ScName, dividend.OnLoad)

	member1 := ctx.NewSoloAgent()
	bal := ctx.Balances(member1)

	dividendMember(ctx, member1, 250)
	require.NoError(t, ctx.Err)

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.Dust - ctx.GasFee
	bal.VerifyBalances(t)

	member2 := ctx.NewSoloAgent()
	bal = ctx.Balances(member1, member2)

	dividendMember(ctx, member2, 500)
	require.NoError(t, ctx.Err)

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.Dust - ctx.GasFee
	bal.VerifyBalances(t)

	member3 := ctx.NewSoloAgent()
	bal = ctx.Balances(member1, member2, member3)

	dividendMember(ctx, member3, 750)
	require.NoError(t, ctx.Err)

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.Dust - ctx.GasFee
	bal.VerifyBalances(t)

	toDivide := 1*iscp.Mi + 1
	dividendDivide(ctx, toDivide)
	require.NoError(t, ctx.Err)

	remain := uint64(toDivide) - toDivide*250/1500 - toDivide*500/1500 - toDivide*750/1500
	bal.Chain += ctx.GasFee
	bal.Originator += remain - ctx.GasFee
	bal.Add(member1, toDivide*250/1500)
	bal.Add(member2, toDivide*500/1500)
	bal.Add(member3, toDivide*750/1500)
	bal.VerifyBalances(t)

	toDivide2 := 2*iscp.Mi + 1
	dividendDivide(ctx, toDivide2)
	require.NoError(t, ctx.Err)

	remain = uint64(toDivide2) - toDivide2*250/1500 - toDivide2*500/1500 - toDivide2*750/1500
	bal.Chain += ctx.GasFee
	bal.Originator += remain - ctx.GasFee
	bal.Add(member1, toDivide2*250/1500)
	bal.Add(member2, toDivide2*500/1500)
	bal.Add(member3, toDivide2*750/1500)
	bal.VerifyBalances(t)
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
