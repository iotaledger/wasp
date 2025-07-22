package models

import "github.com/iotaledger/wasp/v2/packages/parameters"

type NodeOwnerCertificateResponse struct {
	Certificate string `json:"certificate" swagger:"desc(Certificate stating the ownership. (Hex)),required"`
}

type VersionResponse struct {
	Version string `json:"version" swagger:"desc(The version of the node),required"`
}

type InfoResponse struct {
	Version    string               `json:"version" swagger:"desc(The version of the node),required"`
	PublicKey  string               `json:"publicKey" swagger:"desc(The public key of the node (Hex)),required"`
	PeeringURL string               `json:"peeringURL" swagger:"desc(The net id of the node),required"`
	L1Params   *parameters.L1Params `json:"l1Params" swagger:"desc(The L1 parameters),required"`
}
