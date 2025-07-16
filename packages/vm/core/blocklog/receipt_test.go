package blocklog_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestReceiptCodec(t *testing.T) {
	bcs.TestCodec(t, blocklog.RequestReceipt{
		Request: isc.NewOffLedgerRequest(
			isc.EmptyChainID(),
			isc.NewMessage(isc.Hn("0"), isc.Hn("0")),
			123,
			gas.LimitsDefault.MaxGasPerRequest,
		).Sign(cryptolib.NewKeyPair()),
		Error: &isc.UnresolvedVMError{
			ErrorCode: blocklog.ErrBlockNotFound.Code(),
			Params:    []isc.VMErrorParam{uint8(1), uint8(2), "string"},
		},
	})

	bcs.TestCodecAndHash(t, blocklog.RequestReceipt{
		Request: isc.NewOffLedgerRequest(
			isctest.TestChainID,
			isc.NewMessage(isc.Hn("account"), isc.Hn("deposit")),
			123,
			gas.LimitsDefault.MaxGasPerRequest,
		).Sign(cryptolib.TestKeyPair),
		Error: &isc.UnresolvedVMError{
			ErrorCode: blocklog.ErrBlockNotFound.Code(),
			Params:    []isc.VMErrorParam{uint8(1), uint8(2), "string"},
		},
	}, "2e59447923e2")
}

func TestReceiptCodecEVM(t *testing.T) {
	unsignedTx := types.NewTransaction(
		0,
		common.Address{},
		util.Big0,
		100,
		util.Big0,
		[]byte{1, 2, 3},
	)
	ethKey := lo.Must(crypto.GenerateKey())
	tx := lo.Must(types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(42))), ethKey))

	rec := blocklog.RequestReceipt{
		Request: lo.Must(isc.NewEVMOffLedgerTxRequest(
			isctest.RandomChainID(),
			tx,
		)),
	}

	recEnc := bcs.MustMarshal(&rec)
	recDec := bcs.MustUnmarshal[blocklog.RequestReceipt](recEnc)

	// We can't compare the receipts directly because go-ethereum Transaction
	// contains unexporeted time field, which changes internally after RPL encode/decode.
	// So instead we compare string representantion of requests.
	reqStr := rec.Request.String()
	reqDecStr := recDec.Request.String()
	require.Equal(t, reqStr, reqDecStr)
	rec.Request = nil
	recDec.Request = nil
	require.Equal(t, rec, recDec)
}
