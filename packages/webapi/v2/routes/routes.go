// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package routes

func WaitRequestProcessed(chainID, reqID string) string {
	return "/chain/" + chainID + "/request/" + reqID + "/wait"
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
