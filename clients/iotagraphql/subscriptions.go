package iotagraphql

// GetTransactionBlock returns the transaction block from the response if it exists.
func (r TransactionsBySignerWsResponse) GetTransactionBlock() *TransactionsBySignerTransactionsTransactionBlock {
	if r.Data == nil {
		return nil
	}
	txBlock, ok := r.Data.Transactions.(*TransactionsBySignerTransactionsTransactionBlock)
	if !ok {
		return nil
	}
	return txBlock
}

// GetEvent returns the event from the response if it exists.
func (r EventsByModuleWsResponse) GetEvent() *EventsByModuleEventsEvent {
	if r.Data == nil {
		return nil
	}
	event, ok := r.Data.Events.(*EventsByModuleEventsEvent)
	if !ok {
		return nil
	}
	return event
}
