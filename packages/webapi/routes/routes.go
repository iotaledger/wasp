// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package routes

func Info() string {
	return "/info"
}

func NewRequest(chainID string) string {
	return "/request/" + chainID
}

func CallView(chainID, contractHname, functionName string) string {
	return "chain/" + chainID + "/contract/" + contractHname + "/callview/" + functionName
}

func RequestStatus(chainID, reqID string) string {
	return "/chain/" + chainID + "/request/" + reqID + "/status"
}

func WaitRequestProcessed(chainID, reqID string) string {
	return "/chain/" + chainID + "/request/" + reqID + "/wait"
}

func StateGet(chainID, key string) string {
	return "/chain/" + chainID + "/state/" + key
}

func Auth() string {
	return "/adm/auth"
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

func PutCommitteeRecord() string {
	return "/adm/committeerecord"
}

func GetCommitteeRecord(addr string) string {
	return "/adm/committeerecord/" + addr
}

func GetCommitteeForChain(chainID string) string {
	return "/adm/chain/" + chainID + "/committeerecord"
}

func DKSharesPost() string {
	return "/adm/dks"
}

func DKSharesGet(sharedAddress string) string {
	return "/adm/dks/" + sharedAddress
}

func PeeringSelfGet() string {
	return "/adm/peering/self"
}

func PeeringTrustedList() string {
	return "/adm/peering/trusted"
}

func PeeringGetStatus() string {
	return "/adm/peering/established"
}

func PeeringTrustedGet(pubKey string) string {
	return "/adm/peering/trusted/" + pubKey
}

func PeeringTrustedPost() string {
	return PeeringTrustedList()
}

func PeeringTrustedPut(pubKey string) string {
	return PeeringTrustedGet(pubKey)
}

func PeeringTrustedDelete(pubKey string) string {
	return PeeringTrustedGet(pubKey)
}

func Shutdown() string {
	return "/adm/shutdown"
}
