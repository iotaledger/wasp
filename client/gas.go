package client

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func (c *WaspClient) GetGasFeePolicy(chainID isc.ChainID) (*gas.GasFeePolicy, error) {
	res, err := c.CallViewByHname(chainID, governance.Contract.Hname(), governance.ViewGetFeePolicy.Hname(), nil)
	if err != nil {
		return nil, err
	}
	return gas.FeePolicyFromBytes(res.MustGet(governance.ParamFeePolicyBytes))
}
