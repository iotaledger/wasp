// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import "github.com/iotaledger/wasp/client"

type ServiceClient struct {
	waspClient *client.WaspClient
	eventPort  string
}

func NewServiceClient(waspAPI, eventPort string) *ServiceClient {
	return &ServiceClient{waspClient: client.NewWaspClient(waspAPI), eventPort: eventPort}
}

func DefaultServiceClient() *ServiceClient {
	return NewServiceClient("127.0.0.1:9090", "127.0.0.1:5550")
}
