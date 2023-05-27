package rootimpl

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

func eventDeploy(ctx isc.Sandbox, progHash hashing.HashValue, name string, description string) {
	var buf []byte
	buf = append(buf, progHash.Bytes()...)
	buf = append(buf, util.StringToBytes(name)...)
	buf = append(buf, wasmtypes.StringToBytes(description)...)
	ctx.Event("root.deploy", buf)
}

func eventGrant(ctx isc.Sandbox, deployer isc.AgentID) {
	var buf []byte
	buf = append(buf, deployer.Bytes()...)
	ctx.Event("root.grant", buf)
}

func eventRevoke(ctx isc.Sandbox, deployer isc.AgentID) {
	var buf []byte
	buf = append(buf, deployer.Bytes()...)
	ctx.Event("root.revoke", buf)
}
