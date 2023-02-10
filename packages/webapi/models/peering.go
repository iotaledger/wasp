package models

type PeeringNodeStatusResponse struct {
	IsAlive   bool   `json:"isAlive" swagger:"desc(Whether or not the peer is activated),required"`
	NetID     string `json:"netId" swagger:"desc(The NetID of the peer),required"`
	NumUsers  int    `json:"numUsers" swagger:"desc(The amount of users attached to the peer),required"`
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex),required"`
	IsTrusted bool   `json:"isTrusted" swagger:"Desc(Whether or not the peer is trusted),required"`
}

type PeeringNodeIdentityResponse struct {
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex),required"`
	NetID     string `json:"netId" swagger:"desc(The NetID of the peer),required"`
	IsTrusted bool   `json:"isTrusted" swagger:"Desc(Whether or not the peer is trusted),required"`
}

type PeeringNodePublicKeyRequest struct {
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex),required"`
}

type PeeringTrustRequest struct {
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex),required"`
	NetID     string `json:"netId" swagger:"desc(The NetID of the peer),required"`
}
