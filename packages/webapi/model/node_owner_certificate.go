// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

type NodeOwnerCertificateRequest struct {
	NodePubKey   Bytes   `swagger:"desc(Node pub key. (base64))"`
	OwnerAddress Address `swagger:"desc(Node owner address. (base64))"`
}

type NodeOwnerCertificateResponse struct {
	Certificate Bytes `swagger:"desc(Certificate stating the ownership. (base64))"`
}
