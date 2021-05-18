package consensusimpl_old

import "github.com/iotaledger/wasp/packages/coretypes"

func takeIDs(reqs ...coretypes.Request) []coretypes.RequestID {
	ret := make([]coretypes.RequestID, len(reqs))
	for i, req := range reqs {
		ret[i] = req.ID()
	}
	return ret
}

func idsShortStr(ids ...coretypes.RequestID) []string {
	ret := make([]string, len(ids))
	for i, id := range ids {
		ret[i] = id.Short()
	}
	return ret
}
