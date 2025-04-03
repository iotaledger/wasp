package rootimpl

import (
	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
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
