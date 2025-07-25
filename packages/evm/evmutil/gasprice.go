package evmutil

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func CheckGasPrice(gasPrice *big.Int, gasFeePolicy *gas.FeePolicy) error {
	minimumGasPrice := gasFeePolicy.DefaultGasPriceFullDecimals(parameters.BaseTokenDecimals)
	if gasPrice.Cmp(minimumGasPrice) < 0 {
		return fmt.Errorf(
			"insufficient gas price: got %s, minimum is %s",
			gasPrice.Text(10),
			minimumGasPrice.Text(10),
		)
	}
	return nil
}
