package models

type PeeringNodeStatusResponse struct {
	IsAlive   bool   `json:"isAlive" swagger:"desc(Whether or not the peer is activated)"`
	NetID     string `json:"netId" swagger:"desc(The NetID of the peer)"`
	NumUsers  int    `json:"numUsers" swagger:"desc(The amount of users attached to the peer)"`
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex)"`
	IsTrusted bool   `json:"isTrusted" swagger:"Desc(Whether or not the peer is trusted)"`
}

type PeeringNodeIdentityResponse struct {
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex)"`
	NetID     string `json:"netId" swagger:"desc(The NetID of the peer)"`
	IsTrusted bool   `json:"isTrusted" swagger:"Desc(Whether or not the peer is trusted)"`
}

type PeeringNodePublicKeyRequest struct {
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex)"`
}

type PeeringTrustRequest struct {
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex)"`
	NetID     string `json:"netId" swagger:"desc(The NetID of the peer)"`
}
