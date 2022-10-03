package chainimpl

import (
	"time"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/isc"
)

func (c *chainObj) GetAnchorOutput() *isc.AliasOutputWithID {
	return c.stateMgr.GetStatusSnapshot().StateOutput
}

func (c *chainObj) GetTimeData() time.Time {
	return c.consensus.GetStatusSnapshot().TimeData
}

func (c *chainObj) GetDB() kvstore.KVStore {
	return c.db
}
