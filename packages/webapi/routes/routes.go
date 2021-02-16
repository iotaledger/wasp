package routes

func Info() string {
	return "/info"
}

func CallView(contractID string, hname string) string {
	return "/contract/" + contractID + "/callview/" + hname
}

func RequestStatus(chainID string, reqID string) string {
	return "/chain/" + chainID + "/request/" + reqID + "/status"
}

func WaitRequestProcessed(chainID string, reqID string) string {
	return "/chain/" + chainID + "/request/" + reqID + "/wait"
}

func StateQuery(chainID string) string {
	return "/chain/" + chainID + "/state/query"
}

func PutBlob() string {
	return "/blob/put"
}

func GetBlob(hash string) string {
	return "/blob/get/" + hash
}

func HasBlob(hash string) string {
	return "/blob/has/" + hash
}

func ActivateChain(chainID string) string {
	return "/adm/chain/" + chainID + "/activate"
}

func DeactivateChain(chainID string) string {
	return "/adm/chain/" + chainID + "/deactivate"
}

func ListChainRecords() string {
	return "/adm/chainrecords"
}

func PutChainRecord() string {
	return "/adm/chainrecord"
}

func GetChainRecord(chainID string) string {
	return "/adm/chainrecord/" + chainID
}

func DKSharesPost() string {
	return "/adm/dks"
}

func DKSharesGet(sharedAddress string) string {
	return "/adm/dks/" + sharedAddress
}

func DumpState(contractID string) string {
	return "/adm/contract/" + contractID + "/dumpstate"
}

func Shutdown() string {
	return "/adm/shutdown"
}
