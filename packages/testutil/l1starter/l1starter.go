// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package l1starter allows starting and stopping the iota validator tool
// for testing purposes.
package l1starter

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/testcontainers/testcontainers-go"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
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

type Logger struct {
	Prefix string
}

func (l Logger) Printf(s string, args ...interface{}) {
	if l.Prefix != "" {
		fmt.Printf(l.Prefix+": "+s, args...)
	} else {
		fmt.Printf(s, args...)
	}
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

func TestLocal() func() {
	node, cancel := StartNode(context.Background())
	instance.Store(&node)
	return cancel
}

func TestMain(m *testing.M) {
	if instance.Load() != nil {
		m.Run()
		return
	}

	testConfig := LoadConfig()
	var node IotaNodeEndpoint

	if !testConfig.IsLocal {
		iotaNode := NewRemoteIotaNode(testConfig.APIURL, testConfig.FaucetURL, ISCPackageOwner)
		iotaNode.Start(context.Background())

		node = iotaNode
		instance.Store(&node)
	} else {
		var cancel func()
		node, cancel = StartNode(context.Background())
		defer cancel()

		instance.Store(&node)
	}

	rebasedExplorerURL := "https://explorer.rebased.iota.org"
	explorerURL := rebasedExplorerURL + "?network=" + url.QueryEscape(node.APIURL())
	fmt.Printf("L1Starter initialized. \nAPI URL: %s\nFaucet URL: %s\nExplorer URL:%s\n"+
		"(To use local nodes in the Explorer, it is required to disable CORS via an extension or command flags)\n", node.APIURL(), node.FaucetURL(), explorerURL)

	m.Run()
}

func ClusterStart(config L1EndpointConfig) IotaNodeEndpoint {
	if !config.IsLocal {
		iotaNode := NewRemoteIotaNode(config.APIURL, config.FaucetURL, ISCPackageOwner)
		iotaNode.Start(context.Background())

		var iotaNodeEndpoint IotaNodeEndpoint = iotaNode
		instance.Store(&iotaNodeEndpoint)
	} else {
		node, cancel := StartNode(context.Background())
		_ = node // TODO: handle clean up properly
		_ = cancel
		panic("handle clean up properly")
		/* defer cancel()

		instance.Store(&node)
		*/
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
		in.stop(ctx)
	}
}
