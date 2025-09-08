package testutil

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/testutil/testkey"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

const (
	DefaultChainID = uint16(1074)
)

func DummyOffledgerRequest() isc.OffLedgerRequest {
	contract := isc.Hn("somecontract")
	entrypoint := isc.Hn("someentrypoint")
	req := isc.NewOffLedgerRequest(isc.EmptyChainID(), isc.NewMessage(contract, entrypoint), 0, gas.LimitsDefault.MaxGasPerRequest)
	keys, _ := testkey.GenKeyAddr()
	return req.Sign(keys)
}

func DummyOffledgerRequestForAccount(chainID isc.ChainID, nonce uint64, kp *cryptolib.KeyPair) isc.OffLedgerRequest {
	contract := isc.Hn("somecontract")
	entrypoint := isc.Hn("someentrypoint")
	req := isc.NewOffLedgerRequest(chainID, isc.NewMessage(contract, entrypoint), nonce, gas.LimitsDefault.MaxGasPerRequest)
	return req.Sign(kp)
}

func DummyEVMRequest(chainID isc.ChainID, gasPrice *big.Int) isc.OffLedgerRequest {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	tx := types.MustSignNewTx(key, types.NewEIP155Signer(big.NewInt(int64(DefaultChainID))),
		&types.LegacyTx{
			Nonce:    0,
			To:       &common.MaxAddress,
			Value:    big.NewInt(123),
			Gas:      10000,
			GasPrice: gasPrice,
			Data:     []byte{},
		})

	req, err := isc.NewEVMOffLedgerTxRequest(chainID, tx)
	if err != nil {
		panic(err)
	}
	return req
}
