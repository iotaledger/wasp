package models

type PeeringNodeStatus struct {
	IsAlive   bool   `swagger:"desc(Whether or not the peer is activated)"`
	NetID     string `swagger:"desc(The NetID of the peer)"`
	NumUsers  int    `swagger:"desc(The amount of users attached to the peer)"`
	PublicKey string `swagger:"desc(The peers public key encoded in hex)"`
	IsTrusted bool   `swagger:"Desc(Whether or not the peer is trusted)"`
}

type PeeringNodeIdentity struct {
	PublicKey string `swagger:"desc(The peers public key encoded in hex)"`
	NetID     string `swagger:"desc(The NetID of the peer)"`
	IsTrusted bool   `swagger:"Desc(Whether or not the peer is trusted)"`
}

type PeeringNodePublicKey struct {
	PublicKey string `swagger:"desc(The peers public key encoded in hex)"`
}
