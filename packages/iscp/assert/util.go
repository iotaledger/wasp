package assert

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

type Assert struct {
	log iscp.LogInterface
}

func NewAssert(log iscp.LogInterface, name ...string) *Assert {
	return &Assert{
		log: log,
	}
}

func (a Assert) Requiref(cond bool, format string, args ...interface{}) {
	if cond {
		return
	}
	if a.log == nil {
		panic(fmt.Sprintf(format, args...))
	}
	a.log.Panicf(format, args...)
}

func (a Assert) RequireNoError(err error, str ...string) {
	if err != nil {
		if len(str) > 0 {
			panic(xerrors.Errorf("%s: %w", strings.Join(str, " "), err))
		}
		panic(err)
	}
}

func (a Assert) RequireChainOwner(ctx iscp.Sandbox, name ...string) {
	a.RequireCaller(ctx, ctx.ChainOwnerID(), name...)
}

func (a Assert) RequireCaller(ctx iscp.Sandbox, agentID *iscp.AgentID, name ...string) {
	if ctx.Caller().Equals(agentID) {
		return
	}
	if len(name) > 0 {
		a.log.Panicf("%s: unauthorized access", name[0])
	} else {
		a.log.Panicf("unauthorized access")
	}
}
