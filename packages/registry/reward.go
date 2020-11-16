package registry

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/plugins/config"
	flag "github.com/spf13/pflag"
)

const (
	// CfgBindAddress defines the config flag of the web API binding address.
	CfgRewardAddress = "reward.address"
)

func InitFlags() {
	flag.String(CfgRewardAddress, "", "reward address for this Wasp node. Empty (default) means no rewards are collected")
}

func GetFeeDestination(scaddr *address.Address) address.Address {
	//TODO
	ret, err := address.FromBase58(config.Node.GetString(CfgRewardAddress))
	if err != nil {
		ret = address.Address{}
	}
	return ret
}
