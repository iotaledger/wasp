// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package routes

func Info() string {
	return "/info"
}

func NewRequest(chainID string) string {
	return "/chain/" + chainID + "/request"
}

func CallViewByName(chainID, contractHname, functionName string) string {
	return "/chain/" + chainID + "/contract/" + contractHname + "/callview/" + functionName
}

func CallViewByHname(chainID, contractHname, functionHname string) string {
	return "/chain/" + chainID + "/contract/" + contractHname + "/callviewbyhname/" + functionHname
}

func RequestReceipt(chainID, reqID string) string {
	return "/chain/" + chainID + "/request/" + reqID + "/receipt"
}

func WaitRequestProcessed(chainID, reqID string) string {
	return "/chain/" + chainID + "/request/" + reqID + "/wait"
}

func StateGet(chainID, key string) string {
	return "/chain/" + chainID + "/state/" + key
}

func RequestIDByEVMTransactionHash(chainID, txHash string) string {
	return "/chain/" + chainID + "/evm/reqid/" + txHash
}

func EVMJSONRPC(chainID string) string {
	return "/chain/" + chainID + "/evm/jsonrpc"
}

func ActivateChain(chainID string) string {
	return "/chains/" + chainID + "/activate"
}

func DeactivateChain(chainID string) string {
	return "/chains/" + chainID + "/deactivate"
}

func GetChainContracts(chainID string) string {
	return "/chains/" + chainID + "/contracts"
}

func GetChainCommitteeInfo(chainID string) string {
	return "/chains/" + chainID + "/committee"
}

func GetChainInfo(chainID string) string {
	return "/chains/" + chainID + "/info"
}

func GetChainList() string {
	return "/chains"
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

func GetChainsNodeConnectionMetrics() string {
	return "/adm/chains/nodeconn/metrics"
}

func GetChainNodeConnectionMetrics(chainID string) string {
	return "/adm/chains/" + chainID + "/nodeconn/metrics"
}

func GetChainConsensusWorkflowStatus(chainID string) string {
	return "/adm/chains/" + chainID + "/consensus/status"
}

func GetChainConsensusPipeMetrics(chainID string) string {
	return "/adm/chains/" + chainID + "/consensus/metrics/pipe"
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

func Shutdown() string {
	return "/adm/shutdown"
}
