// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

type NodeOwnerCertificateRequest struct {
	NodePubKey   Bytes   `json:"nodePubKey" swagger:"desc(Node pub key. (base64))"`
	OwnerAddress Address `json:"ownerAddress" swagger:"desc(Node owner address. (bech32))"`
}

type NodeOwnerCertificateResponse struct {
	Certificate Bytes `json:"certificate" swagger:"desc(Certificate stating the ownership. (base64))"`
}
