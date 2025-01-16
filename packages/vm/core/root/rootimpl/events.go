package rootimpl

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func eventDeploy(ctx isc.Sandbox, progHash hashing.HashValue, name string) {
	evt := lo.T2(progHash, name)
	ctx.Event("coreroot.deploy", bcs.MustMarshal(&evt))
}

func eventGrant(ctx isc.Sandbox, deployer isc.AgentID) {
	ctx.Event("coreroot.grant", bcs.MustMarshal(&deployer))
}

func eventRevoke(ctx isc.Sandbox, deployer isc.AgentID) {
	ctx.Event("coreroot.revoke", bcs.MustMarshal(&deployer))
}
