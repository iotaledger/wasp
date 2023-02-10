package models

type PeeringNodeStatusResponse struct {
	Name       string `json:"name" swagger:"unique peer name - easy identifier for humans,required"`
	IsAlive    bool   `json:"isAlive" swagger:"desc(Whether or not the peer is activated),required"`
	PeeringURL string `json:"peeringURL" swagger:"desc(The peering URL of the peer),required"`
	NumUsers   int    `json:"numUsers" swagger:"desc(The amount of users attached to the peer),required"`
	PublicKey  string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex),required"`
	IsTrusted  bool   `json:"isTrusted" swagger:"Desc(Whether or not the peer is trusted),required"`
}

type PeeringNodeIdentityResponse struct {
	Name       string `json:"name" swagger:"unique peer name - easy identifier for humans,required"`
	PublicKey  string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex),required"`
	PeeringURL string `json:"peeringURL" swagger:"desc(The peering URL of the peer),required"`
	IsTrusted  bool   `json:"isTrusted" swagger:"Desc(Whether or not the peer is trusted),required"`
}

type PeeringNodePublicKeyRequest struct {
	PublicKey string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex),required"`
}

type PeeringTrustRequest struct {
	Name       string `json:"name" swagger:"unique peer name - easy identifier for humans,required"`
	PublicKey  string `json:"publicKey" swagger:"desc(The peers public key encoded in Hex),required"`
	PeeringURL string `json:"peeringURL" swagger:"desc(The peering URL of the peer),required"`
}
