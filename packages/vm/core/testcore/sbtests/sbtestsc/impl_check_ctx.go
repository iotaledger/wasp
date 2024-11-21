package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/assert"
)

func testCheckContextFromFullEP(ctx isc.Sandbox, chainID isc.ChainID, chainOwnerID isc.AgentID, caller isc.AgentID, agentID isc.AgentID) {
	ctx.Requiref(chainID.Equals(ctx.ChainID()), "fail: chainID")
	ctx.Requiref(chainOwnerID.Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	ctx.Requiref(caller.Equals(ctx.Caller()), "fail: caller")
	myAgentID := isc.NewContractAgentID(ctx.ChainID(), ctx.Contract())
	ctx.Requiref(agentID.Equals(myAgentID), "fail: agentID")
}

func testCheckContextFromViewEP(ctx isc.SandboxView, chainID isc.ChainID, chainOwnerID isc.AgentID, agentID isc.AgentID) {
	a := assert.NewAssert(ctx.Log())

	a.Requiref(chainID.Equals(ctx.ChainID()), "fail: chainID")
	a.Requiref(chainOwnerID.Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	myAgentID := isc.NewContractAgentID(ctx.ChainID(), ctx.Contract())
	a.Requiref(agentID.Equals(myAgentID), "fail: agentID")
}
