package mocknode

import (
	"github.com/iotaledger/goshimmer/packages/txstream/server"
	"github.com/iotaledger/goshimmer/packages/txstream/utxodbledger"
	"github.com/iotaledger/hive.go/logger"
	"go.uber.org/zap"
)

// MockNode provides the bare minimum to emulate a Goshimmer node in a wasp-cluster
// environment, namely the txstream plugin + a few web api endpoints.
type MockNode struct {
	Ledger         *utxodbledger.UtxoDBLedger
	shutdownSignal chan struct{}
	log            *logger.Logger
}

func Start(txStreamBindAddress string, webapiBindAddress string) *MockNode {
	m := &MockNode{
		log:            initLog(),
		Ledger:         utxodbledger.New(),
		shutdownSignal: make(chan struct{}),
	}

	// start the txstream server
	err := server.Listen(m.Ledger, txStreamBindAddress, m.log.Named("txstream"), m.shutdownSignal)
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

func initLog() *logger.Logger {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return log.Sugar()
}
