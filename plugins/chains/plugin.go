package chains

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"sync"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/chain"
	registry_pkg "github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/nodeconn"
)

const PluginName = "Chains"

var (
	log *logger.Logger

	allChains   = make(map[[ledgerstate.AddressLength]byte]chain.Chain)
	chainsMutex = sync.RWMutex{}
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	log.Infof("running %s plugin..", PluginName)
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		if err := initChains(); err != nil {
		}
		// init dispatcher
		chNodeMsg := make(chan interface{}, 100)
		processNodeMsgClosure := events.NewClosure(func(msg interface{}) {
			chNodeMsg <- msg
		})
		go func() {
			for msg := range chNodeMsg {
				processNodeMsg(msg)
			}
		}()
		nodeconn.NodeConn.Events.MessageReceived.Attach(processNodeMsgClosure)

		<-shutdownSignal

		go func() {
			nodeconn.NodeConn.Events.MessageReceived.Detach(processNodeMsgClosure)
			close(chNodeMsg)
		}()
		go shutdownChains()
	})
	if err != nil {
		log.Error(err)
		return
	}
}

func initChains() error {
	chainRecords, err := registry_pkg.GetChainRecords()
	if err != nil {
		return err
	}

	astr := make([]string, len(chainRecords))
	for i := range astr {
		astr[i] = chainRecords[i].ChainID.String()[:10] + ".."
	}
	log.Debugf("loaded %d chain record(s) from registry: %+v", len(chainRecords), astr)

	for _, chr := range chainRecords {
		if chr.Active {
			if err := ActivateChain(chr); err != nil {
				log.Errorf("cannot activate chain %s: %v", chr.ChainID, err)
			}
		}
	}
	return nil
}

func shutdownChains() {
	log.Infof("dismissing all chains...")
	chainsMutex.RLock()
	defer chainsMutex.RUnlock()

	for _, ch := range allChains {
		ch.Dismiss()
	}
	log.Infof("dismissing all chains.. Done")
}
