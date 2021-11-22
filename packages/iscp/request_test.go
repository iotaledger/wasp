package iscp

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

// TODO does not compile
func TestNewRequestData(t *testing.T) {
	t.Run("OnLedger", func(t *testing.T) {
		req, err := OnLedgerFromUTXO(UTXOMetaData{}, &iotago.ExtendedOutput{})
		require.NoError(t, err)
		req.Bytes()
	})
}
