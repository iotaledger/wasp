package nodeconn

import flag "github.com/spf13/pflag"

const (
	CfgNodeAddress = "nodeconn.address"
	CfgNodeAPIBind = "nodeconn.webapi"
)

func init() {
	flag.String(CfgNodeAddress, "127.0.0.1:5000", "node host address")
	flag.String(CfgNodeAPIBind, "127.0.0.1:8080", "webapi bind address")
}
