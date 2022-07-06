package chainimpl

import (
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/iscp"
)

func (ch *chainObj) GetAnchorOutput() *iscp.AliasOutputWithID {
	return ch.stateMgr.GetStatusSnapshot().StateOutput
}

func (ch *chainObj) GetTimeData() time.Time {
	return ch.consensus.GetStatusSnapshot().TimeData
}

func (ch *chainObj) GetDB() kvstore.KVStore {
	return ch.db
}
