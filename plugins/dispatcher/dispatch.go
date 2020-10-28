package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/plugins/chains"
)

func dispatchState(tx *sctransaction.Transaction) {
	txProp := tx.MustProperties() // should be validate while parsing
	if !txProp.IsState() {
		// not state transaction
		return
	}
	cmt := chains.GetChain(*txProp.MustChainID())
	if cmt == nil {
		return
	}
	log.Debugw("dispatchState",
		"txid", tx.ID().String(),
		"chainid", cmt.ID().String(),
	)

	cmt.ReceiveMessage(&chain.StateTransactionMsg{
		Transaction: tx,
	})
}

func dispatchBalances(addr address.Address, bals map[valuetransaction.ID][]*balance.Balance) {
	// pass to the committee by address
	if cmt := chains.GetChain((coretypes.ChainID)(addr)); cmt != nil {
		cmt.ReceiveMessage(chain.BalancesMsg{Balances: bals})
	}
}

func dispatchAddressUpdate(addr address.Address, balances map[valuetransaction.ID][]*balance.Balance, tx *sctransaction.Transaction) {
	log.Debugw("dispatchAddressUpdate", "addr", addr.String())

	cmt := chains.GetChain((coretypes.ChainID)(addr))
	if cmt == nil {
		log.Debugw("committee not found", "addr", addr.String())
		// wrong addressee
		return
	}
	//if _, ok := balances[tx.ID()]; !ok {
	//	// violation of the protocol
	//	log.Errorf("violation of the protocol: transaction %s is not among provided outputs. Ignored", tx.ID().String())
	//	return
	//}
	log.Debugf("received tx with balances: %s", tx.ID().String())

	// update balances before state and requests
	cmt.ReceiveMessage(chain.BalancesMsg{
		Balances: balances,
	})

	txProp := tx.MustProperties() // was parsed before
	if txProp.IsState() && *txProp.MustChainID() == (coretypes.ChainID)(addr) {
		// it is a state update to addr. Send it
		cmt.ReceiveMessage(&chain.StateTransactionMsg{
			Transaction: tx,
		})
	}

	// send all requests to addr
	for i, reqBlk := range tx.Requests() {
		if reqBlk.Target().ChainID() == (coretypes.ChainID)(addr) {
			cmt.ReceiveMessage(&chain.RequestMsg{
				Transaction: tx,
				Index:       coretypes.Uint16(i),
			})
		}
	}
}

func dispatchTxInclusionLevel(level byte, txid *valuetransaction.ID, addrs []address.Address) {
	for _, addr := range addrs {
		cmt := chains.GetChain((coretypes.ChainID)(addr))
		if cmt == nil {
			continue
		}
		cmt.ReceiveMessage(&chain.TransactionInclusionLevelMsg{
			TxId:  txid,
			Level: level,
		})
	}
}
