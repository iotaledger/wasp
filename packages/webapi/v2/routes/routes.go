// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package routes

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

func GetChainsNodeConnectionMetrics() string {
	return "/adm/chains/nodeconn/metrics"
}

func DKSharesPost() string {
	return "/adm/dks"
}

func DKSharesGet(sharedAddress string) string {
	return "/adm/dks/" + sharedAddress
}

func AdmNodeOwnerCertificate() string {
	return "/adm/node/owner/certificate"
}

func Shutdown() string {
	return "/adm/shutdown"
}
