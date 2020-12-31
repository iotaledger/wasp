// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

// This API is used to maintain the distributed key shares.
// The Golang API in this file tries to follow the REST conventions.

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
)

// CreateDKSharesRoute is relative to the AdminRoutePrefix.
func DKSharesPostRoute() string {
	return "dks"
}

// GetDKSharesRoute is relative to the AdminRoutePrefix.
func DKSharesGetRoute(sharedAddress string) string {
	return "dks/" + sharedAddress
}

// DKSharesPostRequest is a POST request for creating new DKShare.
type DKSharesPostRequest struct {
	PeerNetIDs  []string `json:"peerNetIDs" swagger:"desc(NetIDs of the nodes sharing the key.)"`
	PeerPubKeys []string `json:"peerPubKeys" swagger:"desc(Optional, base64 encoded public keys of the peers generating the DKS.)"`
	Threshold   uint16   `json:"threshold" swagger:"desc(Should be =< len(PeerPubKeys))"`
	TimeoutMS   uint16   `json:"timeoutMS" swagger:"desc(Timeout in milliseconds.)"`
}

// DKSharesInfo stands for the DKShare representation, returned by the GET and POST methods.
type DKSharesInfo struct {
	Address      string   `json:"address" swagger:"desc(New generated shared address.)"`
	SharedPubKey string   `json:"sharedPubKey" swagger:"desc(Shared public key (base64-encoded).)"`
	PubKeyShares []string `json:"pubKeyShares" swagger:"desc(Public key shares for all the peers (base64-encoded).)"`
	Threshold    uint16   `json:"threshold"`
	PeerIndex    *uint16  `json:"peerIndex" swagger:"desc(Index of the node returning the share, if it is a member of the sharing group.)"`
}

// DKSharesPost creates new DKShare and returns its state.
func (c *WaspClient) DKSharesPost(request *DKSharesPostRequest) (*DKSharesInfo, error) {
	var response DKSharesInfo
	err := c.do(http.MethodPost, AdminRoutePrefix+"/"+DKSharesPostRoute(), request, &response)
	return &response, err
}

// DKSharesGet retrieves representation of an existing DKShare.
func (c *WaspClient) DKSharesGet(sharedAddress *address.Address) (*DKSharesInfo, error) {
	var sharedAddressStr = sharedAddress.String()
	var response DKSharesInfo
	err := c.do(http.MethodGet, AdminRoutePrefix+"/"+DKSharesGetRoute(sharedAddressStr), nil, &response)
	return &response, err
}
