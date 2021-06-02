package mocknode

import (
	"github.com/iotaledger/goshimmer/packages/txstream/server"
	"github.com/iotaledger/goshimmer/packages/txstream/utxodbledger"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

// MockNode provides the bare minimum to emulate a Goshimmer node in a wasp-cluster
// environment, namely the txstream plugin + a few web api endpoints.
type MockNode struct {
	Ledger         *utxodbledger.UtxoDBLedger
	shutdownSignal chan struct{}
	log            *logger.Logger
}

const debug = false

func Start(txStreamBindAddress string, webapiBindAddress string) *MockNode {
	log := testlogger.NewSimple(debug).Named("txstream")
	log.Infof("starting mocked goshimmer node...")
	m := &MockNode{
		log:            log,
		Ledger:         utxodbledger.New(log),
		shutdownSignal: make(chan struct{}),
	}

	// start the txstream server
	err := server.Listen(m.Ledger, txStreamBindAddress, m.log, m.shutdownSignal)
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
	close(m.shutdownSignal)
}
