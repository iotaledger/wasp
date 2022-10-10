package chainimpl

import (
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

func (c *chainObj) GetAnchorOutput() *isc.AliasOutputWithID {
	return c.stateMgr.GetStatusSnapshot().StateOutput
}

func (c *chainObj) GetTimeData() time.Time {
	return c.consensus.GetStatusSnapshot().TimeData
}

func (c *chainObj) GetVirtualState() (state.VirtualStateAccess, bool, error) {
	return state.LoadSolidState(c.db, c.ID())
}
