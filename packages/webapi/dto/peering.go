// Package dto implements the webapi dto
package dto

import "github.com/iotaledger/wasp/v2/packages/cryptolib"

type PeeringNodeStatus struct {
	Name       string
	IsAlive    bool
	PeeringURL string
	NumUsers   int
	PublicKey  *cryptolib.PublicKey
	IsTrusted  bool
}

type PeeringNodeIdentity struct {
	Name       string
	PublicKey  *cryptolib.PublicKey
	PeeringURL string
	IsTrusted  bool
}
