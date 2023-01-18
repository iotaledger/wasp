package models

import "github.com/iotaledger/wasp/packages/parameters"

type NodeOwnerCertificateRequest struct {
	PublicKey    string `json:"publicKey" swagger:"desc(The public key of the node (Hex))"`
	OwnerAddress string `json:"ownerAddress" swagger:"desc(Node owner address. (Bech32))"`
}

type NodeOwnerCertificateResponse struct {
	Certificate string `json:"certificate" swagger:"desc(Certificate stating the ownership. (Hex))"`
}

type InfoResponse struct {
	Version   string               `json:"version" swagger:"desc(The version of the node)"`
	PublicKey string               `json:"publicKey" swagger:"desc(The public key of the node (Hex))"`
	NetID     string               `json:"netID" swagger:"desc(The net id of the node)"`
	L1Params  *parameters.L1Params `json:"l1Params" swagger:"desc(The L1 parameters)"`
}
