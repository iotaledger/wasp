package evmtypes_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/stretchr/testify/require"
)

func TestTransactionCodec(t *testing.T) {
	tx := types.NewTransaction(
		123,
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		big.NewInt(100),
		100000,
		big.NewInt(100),
		[]byte{1, 2, 3, 4},
	)

	txEnc := bcs.MustMarshal(tx)
	txDec := bcs.MustUnmarshal[*types.Transaction](txEnc)

	require.EqualExportedValues(t, tx, txDec)
}
