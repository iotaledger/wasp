package mocknode

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/txstream"
)

// MockNode provides the bare minimum to emulate a Goshimmer node in a wasp-cluster
// environment, namely the txstream plugin + a few web api endpoints.
type MockNode struct {
	Ledger         *txstream.UtxoDBLedger
	shutdownSignal chan struct{}
	log            *logger.Logger
}

const debug = false

func Start(txStreamBindAddress, webapiBindAddress string) *MockNode {
	log := testlogger.NewSimple(debug).Named("txstream")
	log.Infof("starting mocked goshimmer node...")
	m := &MockNode{
		log:            log,
		Ledger:         txstream.New(log),
		shutdownSignal: make(chan struct{}),
	}

	// start the txstream server
	err := txstream.ServerListen(m.Ledger, txStreamBindAddress, m.log, m.shutdownSignal)
	if err != nil {
		panic(err)
	}

	// start the web api server
	err = m.startWebAPI(webapiBindAddress)
	if err != nil {
		panic(err)
	}

	return m
}

func (m *MockNode) Stop() {
	defer func() {
		err := recover()
		if err != nil {
			m.log.Errorf("recovered from panic while stopping mock node: %v", err) // likely to be caused by test failing + cluster.Stop()
		}
	}()
	close(m.shutdownSignal)
}
