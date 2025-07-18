package evmtypes_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/evm/evmutil"
)

type StructWithTransaction struct {
	Transaction *types.Transaction
	A           int
}

func TestTransactionCodec(t *testing.T) {
	tx := types.NewTransaction(
		123,
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		big.NewInt(100),
		100000,
		big.NewInt(100),
		[]byte{1, 2, 3, 4},
	)

	ethKey := lo.Must(crypto.GenerateKey())
	signedTx := lo.Must(types.SignTx(tx, evmutil.Signer(big.NewInt(int64(42))), ethKey))

	{
		txEnc := bcs.MustMarshal(tx)
		txDec := bcs.MustUnmarshal[*types.Transaction](txEnc)
		require.EqualValues(t, string(lo.Must(tx.MarshalJSON())), string(lo.Must(txDec.MarshalJSON())))
	}
	{
		txEnc := bcs.MustMarshal(signedTx)
		txDec := bcs.MustUnmarshal[*types.Transaction](txEnc)
		require.EqualValues(t, string(lo.Must(signedTx.MarshalJSON())), string(lo.Must(txDec.MarshalJSON())))
	}
	{
		txEnc := bcs.MustMarshal(&StructWithTransaction{
			Transaction: signedTx,
			A:           42,
		})
		txDec := bcs.MustUnmarshal[StructWithTransaction](txEnc)
		require.EqualValues(t, string(lo.Must(signedTx.MarshalJSON())), string(lo.Must(txDec.Transaction.MarshalJSON())))
		require.EqualValues(t, 42, txDec.A)
	}
}
