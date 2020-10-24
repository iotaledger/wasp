package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	_ "github.com/iotaledger/wasp/packages/committee/commiteeimpl" // activate init
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
)

const PluginName = "Dispatcher"

var (
	log *logger.Logger
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
	state.InitLogger()
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {

		chNodeMsg := make(chan interface{}, 100)

		processNodeMsgClosure := events.NewClosure(func(msg interface{}) {
			chNodeMsg <- msg
		})

		processPeerMsgClosure := events.NewClosure(func(msg *peering.PeerMessage) {
			if committee := committees.CommitteeByChainID(msg.ChainID); committee != nil {
				committee.ReceiveMessage(msg)
			}
		})

		err := daemon.BackgroundWorker("wasp dispatcher", func(shutdownSignal <-chan struct{}) {
			// goroutine to read incoming messages from the node
			go func() {
				for msg := range chNodeMsg {
					processNodeMsg(msg)
				}
			}()

			<-shutdownSignal

			log.Infof("Stopping %s..", PluginName)
			go func() {
				nodeconn.EventMessageReceived.Detach(processNodeMsgClosure)
				peering.EventPeerMessageReceived.Detach(processPeerMsgClosure)

				close(chNodeMsg)
				log.Infof("Stopping %s.. Done", PluginName)
			}()
		})
		if err != nil {
			log.Errorf("failed to initialize %v", PluginName)
			return
		}

		// event attachments
		// receiving events from NodeConn --> producing dispatcher events
		nodeconn.EventMessageReceived.Attach(processNodeMsgClosure)
		// receiving messages from peering --> send to respective committees
		peering.EventPeerMessageReceived.Attach(processPeerMsgClosure)

		log.Infof("dispatcher started")

	}, parameters.PriorityDispatcher)

	if err != nil {
		log.Errorf("failed to start worker for %s: %v", PluginName, err)
	}
}

func processNodeMsg(msg interface{}) {
	switch msgt := msg.(type) {

	case *waspconn.WaspFromNodeConfirmedTransactionMsg:
		tx, err := sctransaction.ParseValueTransaction(msgt.Tx)
		if err != nil {
			log.Debugw("!!!! after parsing", "txid", msgt.Tx.ID().String(), "err", err)
			// not a SC transaction. Ignore
			return
		}
		dispatchState(tx)

	case *waspconn.WaspFromNodeAddressOutputsMsg:
		dispatchBalances(msgt.Address, msgt.Balances)

	case *waspconn.WaspFromNodeAddressUpdateMsg:
		tx, err := sctransaction.ParseValueTransaction(msgt.Tx)
		if err != nil {
			log.Debugw("!!!! after parsing", "txid", msgt.Tx.ID().String(), "err", err)
			// not a SC transaction. Ignore
			return
		}
		dispatchAddressUpdate(msgt.Address, msgt.Balances, tx)

	case *waspconn.WaspFromNodeTransactionInclusionLevelMsg:
		dispatchTxInclusionLevel(msgt.Level, &msgt.TxId, msgt.SubscribedAddresses)
	}
}
