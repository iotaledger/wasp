package models

import "github.com/iotaledger/wasp/packages/webapi/v2/types"

type NodeOwnerCertificateRequest struct {
	NodePubKey   types.Base64  `swagger:"desc(Node pub key. (base64))"`
	OwnerAddress types.Address `swagger:"desc(Node owner address. (bech32))"`
}

type NodeOwnerCertificateResponse struct {
	Certificate types.Base64 `swagger:"desc(Certificate stating the ownership. (base64))"`
}
