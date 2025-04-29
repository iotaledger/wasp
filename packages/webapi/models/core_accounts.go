package models

import "github.com/iotaledger/wasp/packages/isc"

type AssetsResponse struct {
	Coins []isc.CoinJSON `json:"coins" swagger:"required"`
}

type AccountObjectsResponse struct {
	ObjectIDs   []string `json:"objectIds" swagger:"required"`
	ObjectTypes []string `json:"objectTypes" swagger:"required"`
}

type AccountNonceResponse struct {
	Nonce string `json:"nonce" swagger:"required,desc(The nonce (uint64 as string))"`
}
