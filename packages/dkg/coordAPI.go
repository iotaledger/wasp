package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This file describes a WebAPI (REST based API) between the DKG
// Coordinator and the DKG nodes (peers generating the shared key).

// TODO: Should all the messages be signed? Move the signature
// handling to the WaspClient and WebAPI?

// The overall procedure is the following (from the coordinator's point of view):
//
// DkgInit(addrs, InitReq) -*->
//     <-*- OK
//
// DkgStep(addrs, "sendDeals")
//     <-*- OK
//
// ...
//
// DkgPubKey(addrs)
//     <-*- PubKeyResp
//     Verify all pub keys equal.
//
// DkgStep(addrs, "commit")
//     <-*- OK

// CoordNodeProvider is used by the coordinator to access the
// DKG nodes. This abstraction is used to replace the WebClient
// with a mock in the unit tests.
type CoordNodeProvider interface {
	DkgInit(peerAddrs []string, dkgID string, msg *InitReq) error      // PUT  /dkg/:DkgID
	DkgStep(peerAddrs []string, dkgID string, msg *StepReq) error      // POST /dkg/:DkgID/step/
	DkgPubKey(peerAddrs []string, dkgID string) ([]*PubKeyResp, error) // GET  /dkg/:DkgID
}

// InitReq is a message sent by the coordinator to all the peers to
// initiate the DKG process.
type InitReq struct {
	PeerLocs  []string `json:"peerLocs"`
	PeerPubs  [][]byte `json:"peerPubs"`
	CoordPub  []byte   `json:"coordPub"`
	Treshold  uint32   `json:"treshold"`
	Version   byte     `json:"version"`
	TimeoutMS uint64   `json:"timeoutMS"`
}

// StepReq is a message used to coordinate the DKG procedure by
// ensuring the lock-step, as required by the DKG algorithm
// assumptions (Rabin as well as Peddersen).
type StepReq struct {
	Step string `json:"step"`
}

// PubKeyResp is a message responded to the coodinator
// by the DKG peers returning the shared public key.
// All the nodes must return the same public key.
type PubKeyResp struct {
	ChainID   []byte `json:"address"`
	PubKey    []byte `json:"pubKey"`
	Signature []byte `json:"signature"`
}
