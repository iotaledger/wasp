// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// l1starter allows starting and stopping the iota validator tool
// for testing purposes.
package l1starter

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/testcontainers/testcontainers-go"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

var (
	ISCPackageOwner iotasigner.Signer
	instance        atomic.Pointer[IotaNodeEndpoint]
)

type Ports struct {
	RPC    int
	Faucet int
}

type Config struct {
	Host    string
	Ports   Ports
	Logger  testcontainers.Logging
	TempDir string
}

type Logger struct{}

func (l Logger) Printf(s string, args ...interface{}) {
	fmt.Printf(s, args...)
}

type IotaNodeEndpoint interface {
	ISCPackageID() iotago.PackageID
	APIURL() string
	FaucetURL() string
	L1Client() clients.L1Client
	IsLocal() bool
}

func init() {
	var seed [ed25519.SeedSize]byte = cryptolib.NewSeed()
	ISCPackageOwner = iotasigner.NewSigner(seed[:], iotasigner.KeySchemeFlagDefault)
}

func Instance() IotaNodeEndpoint {
	in := instance.Load()
	if in == nil {
		panic("LocalIotaNode not started; call Start() first")
	}
	return *in
}

func IsLocalConfigured() bool {
	testConfig := LoadConfig()
	return testConfig.IsLocal
}

func TestMain(m *testing.M) {
	if instance.Load() != nil {
		m.Run()
		return
	}

	testConfig := LoadConfig()

	if !testConfig.IsLocal {
		iotaNode := NewRemoteIotaNode(testConfig.APIURL, testConfig.FaucetURL, ISCPackageOwner)
		iotaNode.start(context.Background())

		var iotaNodeEndpoint IotaNodeEndpoint = iotaNode
		instance.Store(&iotaNodeEndpoint)
	} else {
		node, cancel := StartNode(context.Background())
		defer cancel()

		instance.Store(&node)
	}

	m.Run()
}

func ClusterStart(config L1EndpointConfig) IotaNodeEndpoint {
	if !config.IsLocal {
		iotaNode := NewRemoteIotaNode(config.APIURL, config.FaucetURL, ISCPackageOwner)
		iotaNode.start(context.Background())

		var iotaNodeEndpoint IotaNodeEndpoint = iotaNode
		instance.Store(&iotaNodeEndpoint)
	} else {
		node, cancel := StartNode(context.Background())
		panic("handle clean up properly")
		defer cancel()

		instance.Store(&node)
	}

	return *instance.Load()
}

func ISCPackageID() iotago.PackageID {
	return Instance().ISCPackageID()
}

func StartNode(ctx context.Context) (IotaNodeEndpoint, func()) {
	in := NewLocalIotaNode(ISCPackageOwner)

	in.start(ctx)

	return in, func() {
		in.stop()
	}
}
