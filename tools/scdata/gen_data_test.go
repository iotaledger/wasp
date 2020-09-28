package main

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"testing"
)

func TestGenData(t *testing.T) {
	t.Logf("dummy color: %s", sctransaction.RandomColor().String())
	t.Logf("dummy owner's address: %s", address.RandomOfType(address.VersionBLS).String())
	t.Logf("dummy program hash: %s", hashing.HashStrings("dummy program").String())
}
