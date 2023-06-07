package rootimpl

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

func eventDeploy(ctx isc.Sandbox, progHash hashing.HashValue, name string, description string) {
	w := new(bytes.Buffer)
	_ = util.WriteN(w, progHash.Bytes())
	_ = util.WriteString16(w, name)
	_ = util.WriteString16(w, description)
	ctx.Event("coreroot.deploy", w.Bytes())
}

func eventGrant(ctx isc.Sandbox, deployer isc.AgentID) {
	w := new(bytes.Buffer)
	_ = util.WriteN(w, deployer.Bytes())
	ctx.Event("coreroot.grant", w.Bytes())
}

func eventRevoke(ctx isc.Sandbox, deployer isc.AgentID) {
	w := new(bytes.Buffer)
	_ = util.WriteN(w, deployer.Bytes())
	ctx.Event("coreroot.revoke", w.Bytes())
}
