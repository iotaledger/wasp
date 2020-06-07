package nodeconn

import flag "github.com/spf13/pflag"

const (
	CfgNodeAddress = "nodeconn.address"
)

func init() {
	flag.String(CfgNodeAddress, "127.0.0.1:5000", "node host address")
}
