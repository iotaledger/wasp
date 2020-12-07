package scclient

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

func (c *SCClient) PostRequest(
	fname string,
	mint map[address.Address]int64,
	transfer map[balance.Color]int64,
	args dict.Dict,
) (*sctransaction.Transaction, error) {
	return c.ChainClient.PostRequest(
		c.ContractHname,
		coretypes.Hn(fname),
		mint,
		transfer,
		args,
	)
}
