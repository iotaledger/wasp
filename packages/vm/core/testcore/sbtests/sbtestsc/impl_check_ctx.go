package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/assert"
)

func testCheckContextFromFullEP(ctx isc.Sandbox, chainAdmin isc.AgentID, caller isc.AgentID, agentID isc.AgentID) {
	ctx.Requiref(chainAdmin.Equals(ctx.ChainAdmin()), "fail: chainAdmin")
	ctx.Requiref(caller.Equals(ctx.Caller()), "fail: caller")
	myAgentID := isc.NewContractAgentID(ctx.Contract())
	ctx.Requiref(agentID.Equals(myAgentID), "fail: agentID")
}

func testCheckContextFromViewEP(ctx isc.SandboxView, chainAdmin isc.AgentID, agentID isc.AgentID) {
	a := assert.NewAssert(ctx.Log())

	a.Requiref(chainAdmin.Equals(ctx.ChainAdmin()), "fail: chainAdmin")
	myAgentID := isc.NewContractAgentID(ctx.Contract())
	a.Requiref(agentID.Equals(myAgentID), "fail: agentID")
}
