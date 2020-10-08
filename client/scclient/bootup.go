package scclient

import (
	"strings"

	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/registry"
)

func (sc *SCClient) BootupData() (*registry.BootupData, error) {
	// TODO move here
	b, _, err := apilib.GetSCData(strings.Replace(sc.WaspClient.BaseURL(), "http://", "", 1), sc.Address)
	return b, err
}
