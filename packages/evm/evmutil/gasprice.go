package evmutil

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func CheckGasPrice(tx *types.Transaction, gasFeePolicy *gas.FeePolicy) error {
	expectedGasPrice := gasFeePolicy.GasPriceWei(parameters.L1().BaseToken.Decimals)
	gasPrice := tx.GasPrice()
	if gasPrice.Cmp(expectedGasPrice) != 0 {
		return fmt.Errorf(
			"invalid gas price: got %s, want %s",
			gasPrice.Text(10),
			expectedGasPrice.Text(10),
		)
	}
	return nil
}
