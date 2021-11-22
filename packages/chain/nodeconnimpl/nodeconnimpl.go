package nodeconnimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/txstream"
	txstream_client "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
)

type NodeConnImplementation struct {
	client                 *txstream_client.Client
	transactionHandlers    map[ledgerstate.AliasAddress]chain.NodeConnectionHandleTransactionFun
	iStateHandlers         map[ledgerstate.AliasAddress]chain.NodeConnectionHandleInclusionStateFun
	outputHandlers         map[ledgerstate.AliasAddress]chain.NodeConnectionHandleOutputFun
	unspentAOutputHandlers map[ledgerstate.AliasAddress]chain.NodeConnectionHandleUnspentAliasOutputFun
	log                    *logger.Logger // general chains logger
}

var _ chain.NodeConnection = &NodeConnImplementation{}

func NewNodeConnection(nodeConnClient *txstream_client.Client, log *logger.Logger) chain.NodeConnection {
	ret := &NodeConnImplementation{
		client:                 nodeConnClient,
		transactionHandlers:    make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleTransactionFun),
		iStateHandlers:         make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleInclusionStateFun),
		outputHandlers:         make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleOutputFun),
		unspentAOutputHandlers: make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleUnspentAliasOutputFun),
		log:                    log,
	}
	ret.client.Events.TransactionReceived.Attach(events.NewClosure(func(msg *txstream.MsgTransaction) {
		ret.log.Debugf("NodeConnnection::TransactionReceived...")
		defer ret.log.Debugf("NodeConnnection::TransactionReceived... Done")
		aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
		if !ok {
			ret.log.Warnf("NodeConnnection::TransactionReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
			return
		}
		handler, ok := ret.transactionHandlers[*aliasAddr]
		if !ok {
			ret.log.Warnf("NodeConnnection::TransactionReceived: no handler for address %v", aliasAddr.String())
			return
		}
		handler(msg.Tx)
	}))
	ret.client.Events.InclusionStateReceived.Attach(events.NewClosure(func(msg *txstream.MsgTxInclusionState) {
		ret.log.Debugf("NodeConnnection::InclusionStateReceived...")
		defer ret.log.Debugf("NodeConnnection::InclusionStateReceived... Done")
		aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
		if !ok {
			ret.log.Warnf("NodeConnnection::InclusionStateReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
			return
		}
		handler, ok := ret.iStateHandlers[*aliasAddr]
		if !ok {
			ret.log.Warnf("NodeConnnection::InclusionStateReceived: no handler for address %v", aliasAddr.String())
			return
		}
		handler(msg.TxID, msg.State)
	}))
	ret.client.Events.OutputReceived.Attach(events.NewClosure(func(msg *txstream.MsgOutput) {
		ret.log.Debugf("NodeConnnection::OutputReceived...")
		defer ret.log.Debugf("NodeConnnection::OutputReceived... Done")
		aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
		if !ok {
			ret.log.Warnf("NodeConnnection::OutputReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
			return
		}
		handler, ok := ret.outputHandlers[*aliasAddr]
		if !ok {
			ret.log.Warnf("NodeConnnection::OutputReceived: no handler for address %v", aliasAddr.String())
			return
		}
		handler(msg.Output)
	}))
	ret.client.Events.UnspentAliasOutputReceived.Attach(events.NewClosure(func(msg *txstream.MsgUnspentAliasOutput) {
		ret.log.Debugf("NodeConnnection::UnspentAliasOutputReceived...")
		defer ret.log.Debugf("NodeConnnection::UnspentAliasOutputReceived... Done")
		handler, ok := ret.unspentAOutputHandlers[*msg.AliasAddress]
		if !ok {
			ret.log.Warnf("NodeConnnection::UnspentAliasOutputReceived: no handler for address %v", msg.AliasAddress.String())
			return
		}
		handler(msg.AliasOutput, msg.Timestamp)
	}))
	return ret
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

func (n *NodeConnImplementation) AttachToTransactionReceived(addr *ledgerstate.AliasAddress, handler chain.NodeConnectionHandleTransactionFun) {
	n.transactionHandlers[*addr] = handler
}

func (n *NodeConnImplementation) AttachToInclusionStateReceived(addr *ledgerstate.AliasAddress, handler chain.NodeConnectionHandleInclusionStateFun) {
	n.iStateHandlers[*addr] = handler
}

func (n *NodeConnImplementation) AttachToOutputReceived(addr *ledgerstate.AliasAddress, handler chain.NodeConnectionHandleOutputFun) {
	n.outputHandlers[*addr] = handler
}

func (n *NodeConnImplementation) AttachToUnspentAliasOutputReceived(addr *ledgerstate.AliasAddress, handler chain.NodeConnectionHandleUnspentAliasOutputFun) {
	n.unspentAOutputHandlers[*addr] = handler
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
