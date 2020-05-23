package state

import (
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/mr-tron/base58"
)

func dbKeyProcessedRequest(reqId *sctransaction.RequestId) []byte {
	return reqId.Bytes()
}

func MarkRequestsProcessed(reqids []*sctransaction.RequestId) error {
	dbase, err := database.GetProcessedRequestsDB()
	if err != nil {
		return err
	}

	for _, rid := range reqids {
		key := dbKeyProcessedRequest(rid)

		log.Debugw("request marking processed in db",
			"key", base58.Encode(key),
			"reqid", rid.String(),
		)
		err := dbase.Set(database.Entry{
			Key: key,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func IsRequestProcessed(reqid *sctransaction.RequestId) (bool, error) {
	dbase, err := database.GetProcessedRequestsDB()
	if err != nil {
		return false, err
	}
	contains, err := dbase.Contains(dbKeyProcessedRequest(reqid))
	if err != nil {
		return false, err
	}
	return contains, nil
}
