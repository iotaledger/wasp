package state

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/labstack/gommon/log"
)

// assumes it is consistent, just save into the db
// must an atomic call to badger TODO
func SaveStateToDb(stateUpd StateUpdate, varState VariableState, reqIds *[]sctransaction.RequestId) error {
	if err := varState.SaveToDb(); err != nil {
		return err
	}
	if err := stateUpd.SaveToDb(); err != nil {
		// this is very bad!!!!
		return err
	}

	if err := MarkReqIdsProcessed(reqIds); err != nil {
		// very bad
		return err
	}
	// ---- end of must be atomic. how to make it?
	log.Infof("variable state #%d has been solidified for sc addr %s", varState.StateIndex(), stateUpd.Address().String())
	return nil
}

func MarkReqIdsProcessed(reqIds *[]sctransaction.RequestId) error {
	for _, reqid := range *reqIds {
		if err := MarkRequestProcessed(&reqid); err != nil {
			return err
		}
	}
	return nil
}
