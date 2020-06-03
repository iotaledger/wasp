package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	_ "github.com/iotaledger/wasp/packages/committee/commiteeimpl" // activate init
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
	"time"
)

// PluginName is the name of the database plugin.
const PluginName = "Dispatcher"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
	state.InitLogger()
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {

		chNodeMsgData := make(chan []byte, 100)

		processNodeDataClosure := events.NewClosure(func(data []byte) {
			chNodeMsgData <- data
		})

		processPeerMsgClosure := events.NewClosure(func(msg *peering.PeerMessage) {
			if committee := committees.CommitteeByAddress(msg.Address); committee != nil {
				committee.ReceiveMessage(msg)
			}
		})

		err := daemon.BackgroundWorker("wasp dispatcher", func(shutdownSignal <-chan struct{}) {
			// goroutine to read incoming messages from the node
			go func() {
				for data := range chNodeMsgData {
					processNodeMsgData(data)
				}
			}()

			<-shutdownSignal

			log.Infof("Stopping %s..", PluginName)
			go func() {
				nodeconn.EventNodeMessageReceived.Detach(processNodeDataClosure)
				peering.EventPeerMessageReceived.Detach(processPeerMsgClosure)
				Events.OutputsArrivedFromNode.DetachAll()
				Events.TransactionArrivedFromNode.DetachAll()

				close(chNodeMsgData)
				log.Infof("Stopping %s.. Done", PluginName)
			}()
		})
		if err != nil {
			log.Errorf("failed to initialize %v", PluginName)
			return
		}

		// event attachments
		// receiving events from NodeConn --> producing dispatcher events
		nodeconn.EventNodeMessageReceived.Attach(processNodeDataClosure)
		// receiving messages from peering --> send to respective committees
		peering.EventPeerMessageReceived.Attach(processPeerMsgClosure)

		// dispatcher events. It is consumed by dispatcher but other parts may attach too
		// when transaction arrives from node
		Events.TransactionArrivedFromNode.Attach(events.NewClosure(func(tx *sctransaction.Transaction) {
			dispatchState(tx)
		}))
		// when balances arrive from nodes
		Events.OutputsArrivedFromNode.Attach(events.NewClosure(func(addr address.Address, balances map[valuetransaction.ID][]*balance.Balance) {
			dispatchBalances(addr, balances)
		}))
		Events.AddressUpdateArrivedFromNode.Attach(events.NewClosure(func(addr address.Address, balances map[valuetransaction.ID][]*balance.Balance, tx *sctransaction.Transaction) {
			dispatchAddressUpdate(addr, balances, tx)
		}))

		log.Infof("dispatcher started")

	}, shutdown.PriorityDispatcher)

	if err != nil {
		log.Errorf("failed to start worker for %s: %v", PluginName, err)
	}
}

func processNodeMsgData(data []byte) {
	msg, err := waspconn.DecodeMsg(data, true)
	if err != nil {
		log.Errorf("wrong message from node: %v", err)
		return
	}

	log.Debugf("received msg type %T data len = %d", msg, len(data))

	switch msgt := msg.(type) {
	case *waspconn.WaspPingMsg:
		roundtrip := time.Since(time.Unix(0, msgt.Timestamp))
		log.Infof("PING %d response from node. Roundtrip %v", msgt.Id, roundtrip)

	case *waspconn.WaspFromNodeTransactionMsg:
		tx, err := sctransaction.ParseValueTransaction(msgt.Tx)
		if err != nil {
			log.Debugw("!!!! after parsing", "txid", msgt.Tx.ID().String(), "err", err)
			// not a SC transaction. Ignore
			return
		}
		Events.TransactionArrivedFromNode.Trigger(tx)

	case *waspconn.WaspFromNodeAddressOutputsMsg:
		Events.OutputsArrivedFromNode.Trigger(msgt.Address, msgt.Balances)

	case *waspconn.WaspFromNodeAddressUpdateMsg:
		tx, err := sctransaction.ParseValueTransaction(msgt.Tx)
		if err != nil {
			log.Debugw("!!!! after parsing", "txid", msgt.Tx.ID().String(), "err", err)
			// not a SC transaction. Ignore
			return
		}
		Events.AddressUpdateArrivedFromNode.Trigger(msgt.Address, msgt.Balances, tx)
	}
}
