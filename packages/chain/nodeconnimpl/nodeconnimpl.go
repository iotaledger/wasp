package nodeconnimpl

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/txstream"
	txstream_client "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
)

type NodeConnImplementation struct {
	client *txstream_client.Client
	log    *logger.Logger // general chains logger
}

var _ chain.NodeConnection = &NodeConnImplementation{}

func NewNodeConnection(nodeConnClient *txstream_client.Client, log *logger.Logger) chain.NodeConnection {
	return &NodeConnImplementation{
		client: nodeConnClient,
		log:    log,
	}
}

// NOTE: NodeConnectionSender methods are logged through each chain logger in ChainNodeConnImplementation

func (n *NodeConnImplementation) PullState(addr *ledgerstate.AliasAddress) {
	n.client.RequestUnspentAliasOutput(addr)
}

func (n *NodeConnImplementation) PullTransactionInclusionState(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	n.client.RequestTxInclusionState(addr, txid)
}

func (n *NodeConnImplementation) PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
	n.client.RequestConfirmedOutput(addr, outputID)
}

func (n *NodeConnImplementation) PostTransaction(tx *ledgerstate.Transaction) {
	n.client.PostTransaction(tx)
}

func (n *NodeConnImplementation) AttachToTransactionReceived(fun func(*ledgerstate.AliasAddress, *ledgerstate.Transaction)) {
	n.client.Events.TransactionReceived.Attach(events.NewClosure(func(msg *txstream.MsgTransaction) {
		n.log.Debugf("NodeConnnection::TransactionReceived...")
		defer n.log.Debugf("NodeConnnection::TransactionReceived... Done")
		aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
		if !ok {
			n.log.Warnf("NodeConnnection::TransactionReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
			return
		}
		fun(aliasAddr, msg.Tx)
	}))
}

func (n *NodeConnImplementation) AttachToInclusionStateReceived(fun func(*ledgerstate.AliasAddress, ledgerstate.TransactionID, ledgerstate.InclusionState)) {
	n.client.Events.InclusionStateReceived.Attach(events.NewClosure(func(msg *txstream.MsgTxInclusionState) {
		n.log.Debugf("NodeConnnection::InclusionStateReceived...")
		defer n.log.Debugf("NodeConnnection::InclusionStateReceived... Done")
		aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
		if !ok {
			n.log.Warnf("NodeConnnection::InclusionStateReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
			return
		}
		fun(aliasAddr, msg.TxID, msg.State)
	}))
}

func (n *NodeConnImplementation) AttachToOutputReceived(fun func(*ledgerstate.AliasAddress, ledgerstate.Output)) {
	n.client.Events.OutputReceived.Attach(events.NewClosure(func(msg *txstream.MsgOutput) {
		n.log.Debugf("NodeConnnection::OutputReceived...")
		defer n.log.Debugf("NodeConnnection::OutputReceived... Done")
		aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
		if !ok {
			n.log.Warnf("NodeConnnection::OutputReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
			return
		}
		fun(aliasAddr, msg.Output)
	}))
}

func (n *NodeConnImplementation) AttachToUnspentAliasOutputReceived(fun func(*ledgerstate.AliasAddress, *ledgerstate.AliasOutput, time.Time)) {
	n.client.Events.UnspentAliasOutputReceived.Attach(events.NewClosure(func(msg *txstream.MsgUnspentAliasOutput) {
		n.log.Debugf("NodeConnnection::UnspentAliasOutputReceived...")
		defer n.log.Debugf("NodeConnnection::UnspentAliasOutputReceived... Done")
		fun(msg.AliasAddress, msg.AliasOutput, msg.Timestamp)
	}))
}

func (n *NodeConnImplementation) Subscribe(addr ledgerstate.Address) {
	n.log.Debugf("NodeConnnection::Subscribing to %v...", addr.String())
	defer n.log.Debugf("NodeConnnection::Subscribing done")
	n.client.Subscribe(addr)
}

func (n *NodeConnImplementation) Unsubscribe(addr ledgerstate.Address) {
	n.log.Debugf("NodeConnnection::Unsubscribing from %v...", addr.String())
	defer n.log.Debugf("NodeConnnection::Unsubscribing done")
	n.client.Unsubscribe(addr)
}
