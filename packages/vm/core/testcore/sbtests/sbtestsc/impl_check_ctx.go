package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/assert"
)

func testCheckContextFromFullEP(ctx isc.Sandbox, chainOwnerID isc.AgentID, caller isc.AgentID, agentID isc.AgentID) {
	ctx.Requiref(chainOwnerID.Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	ctx.Requiref(caller.Equals(ctx.Caller()), "fail: caller")
	myAgentID := isc.NewContractAgentID(ctx.Contract())
	ctx.Requiref(agentID.Equals(myAgentID), "fail: agentID")
}

func testCheckContextFromViewEP(ctx isc.SandboxView, chainOwnerID isc.AgentID, agentID isc.AgentID) {
	a := assert.NewAssert(ctx.Log())

	a.Requiref(chainOwnerID.Equals(ctx.ChainOwnerID()), "fail: chainOwnerID")
	myAgentID := isc.NewContractAgentID(ctx.Contract())
	a.Requiref(agentID.Equals(myAgentID), "fail: agentID")
}
