package fa

import (
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/spf13/pflag"
)

var Config = &config.SCConfig{
	ShortName: "fa",
	Flags:     pflag.NewFlagSet("fairauction", pflag.ExitOnError),
}
