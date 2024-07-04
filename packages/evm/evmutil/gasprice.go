package evmutil

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func CheckGasPrice(gasPrice *big.Int, gasFeePolicy *gas.FeePolicy) error {
	minimumGasPrice := gasFeePolicy.DefaultGasPriceFullDecimals(parameters.Decimals)
	if gasPrice.Cmp(minimumGasPrice) < 0 {
		return fmt.Errorf(
			"insufficient gas price: got %s, minimum is %s",
			gasPrice.Text(10),
			minimumGasPrice.Text(10),
		)
	}
	return nil
}
