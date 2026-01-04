package graphqltypes

import "github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

// TransactionBytes represents a built transaction ready to be signed and executed
// This is a simplified version without nested wrapper types
type TransactionBytes struct {
	// The gas object to be used
	Gas []*iotago.ObjectRef

	// Objects to be used in this transaction
	InputObjects []map[string]interface{}

	// Transaction data bytes
	TxBytes []byte
}
