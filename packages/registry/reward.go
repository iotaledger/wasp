package registry

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	// "github.com/iotaledger/wasp/packages/parameters"
	flag "github.com/spf13/pflag"
)

const (
	// CfgBindAddress defines the config flag of the web API binding address.
	CfgRewardAddress = "reward.address"
)

func InitFlags() {
	flag.String(CfgRewardAddress, "", "reward address for this Wasp node. Empty (default) means no rewards are collected")
}

func GetFeeDestination(chainID *iscp.ChainID) iotago.Address {
	// TODO
	/*ret, err := iotago.AddressFromBase58EncodedString(parameters.GetString(CfgRewardAddress))
	if err != nil {
		return nil
	}
	return ret*/
	return nil
}
