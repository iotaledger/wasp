package dto

import "github.com/iotaledger/wasp/packages/cryptolib"

type PeeringNodeStatus struct {
	IsAlive   bool
	NetID     string
	NumUsers  int
	PublicKey *cryptolib.PublicKey
	IsTrusted bool
}

type PeeringNodeIdentity struct {
	PublicKey *cryptolib.PublicKey
	NetID     string
	IsTrusted bool
}
