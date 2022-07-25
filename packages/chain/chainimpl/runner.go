package chainimpl

import (
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/iscp"
)

func (c *chainObj) GetAnchorOutput() *iscp.AliasOutputWithID {
	return c.stateMgr.GetStatusSnapshot().StateOutput
}

func (c *chainObj) GetTimeData() time.Time {
	return c.consensus.GetStatusSnapshot().TimeData
}

func (c *chainObj) GetDB() kvstore.KVStore {
	return c.db
}
