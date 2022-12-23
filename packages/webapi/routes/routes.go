// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package routes

func Info() string {
	return "/info"
}

func NewRequest(chainIDBech32 string) string {
	return "/chain/" + chainIDBech32 + "/request"
}

func CallViewByName(chainIDBech32, contractHname, functionName string) string {
	return "/chain/" + chainIDBech32 + "/contract/" + contractHname + "/callview/" + functionName
}

func CallViewByHname(chainIDBech32, contractHname, functionHname string) string {
	return "/chain/" + chainIDBech32 + "/contract/" + contractHname + "/callviewbyhname/" + functionHname
}

func RequestReceipt(chainIDBech32, reqID string) string {
	return "/chain/" + chainIDBech32 + "/request/" + reqID + "/receipt"
}

func WaitRequestProcessed(chainIDBech32, reqID string) string {
	return "/chain/" + chainIDBech32 + "/request/" + reqID + "/wait"
}

func StateGet(chainIDBech32, key string) string {
	return "/chain/" + chainIDBech32 + "/state/" + key
}

func RequestIDByEVMTransactionHash(chainIDBech32, txHash string) string {
	return "/chain/" + chainIDBech32 + "/evm/reqid/" + txHash
}

func EVMJSONRPC(chainIDBech32 string) string {
	return "/chain/" + chainIDBech32 + "/evm/jsonrpc"
}

func ActivateChain(chainIDBech32 string) string {
	return "/adm/chain/" + chainIDBech32 + "/activate"
}

func DeactivateChain(chainIDBech32 string) string {
	return "/adm/chain/" + chainIDBech32 + "/deactivate"
}

func GetChainInfo(chainIDBech32 string) string {
	return "/adm/chain/" + chainIDBech32 + "/info"
}

func ListChainRecords() string {
	return "/adm/chainrecords"
}

func PutChainRecord() string {
	return "/adm/chainrecord"
}

func GetChainRecord(chainIDBech32 string) string {
	return "/adm/chainrecord/" + chainIDBech32
}

func GetChainsNodeConnectionMetrics() string {
	return "/adm/chain/nodeconn/metrics"
}

func GetChainNodeConnectionMetrics(chainIDBech32 string) string {
	return "/adm/chain/" + chainIDBech32 + "/nodeconn/metrics"
}

func GetChainConsensusWorkflowStatus(chainIDBech32 string) string {
	return "/adm/chain/" + chainIDBech32 + "/consensus/status"
}

func GetChainConsensusPipeMetrics(chainIDBech32 string) string {
	return "/adm/chain/" + chainIDBech32 + "/consensus/metrics/pipe"
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

func AdmNodeOwnerCertificate() string {
	return "/adm/node/owner/certificate"
}

func AdmAddAccessNode(chainIDBech32 string) string {
	return "/adm/chain/" + chainIDBech32 + "/access-node/add"
}

func AdmRemoveAccessNode(chainIDBech32 string) string {
	return "/adm/chain/" + chainIDBech32 + "/access-node/remove"
}

func Shutdown() string {
	return "/adm/shutdown"
}
