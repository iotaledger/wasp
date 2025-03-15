package testutil

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func DummyOffledgerRequest() isc.OffLedgerRequest {
	contract := isc.Hn("somecontract")
	entrypoint := isc.Hn("someentrypoint")
	req := isc.NewOffLedgerRequest(isc.NewMessage(contract, entrypoint), 0, gas.LimitsDefault.MaxGasPerRequest)
	keys, _ := testkey.GenKeyAddr()
	return req.Sign(keys)
}

func DummyOffledgerRequestForAccount(nonce uint64, kp *cryptolib.KeyPair) isc.OffLedgerRequest {
	contract := isc.Hn("somecontract")
	entrypoint := isc.Hn("someentrypoint")
	req := isc.NewOffLedgerRequest(isc.NewMessage(contract, entrypoint), nonce, gas.LimitsDefault.MaxGasPerRequest)
	return req.Sign(kp)
}

func DummyEVMRequest(gasPrice *big.Int) isc.OffLedgerRequest {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	tx := types.MustSignNewTx(key, types.NewEIP155Signer(big.NewInt(0)),
		&types.LegacyTx{
			Nonce:    0,
			To:       &common.MaxAddress,
			Value:    big.NewInt(123),
			Gas:      10000,
			GasPrice: gasPrice,
			Data:     []byte{},
		})

	req, err := isc.NewEVMOffLedgerTxRequest(tx)
	if err != nil {
		panic(err)
	}
	return req
}
