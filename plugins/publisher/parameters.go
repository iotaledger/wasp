package publisher

import (
	flag "github.com/spf13/pflag"
)

const (
	// CfgBindAddress defines the config flag of the web API binding address.
	CfgNanomsgPublisherPort = "nanomsg.port"
)

func init() {
	flag.Int(CfgNanomsgPublisherPort, 5550, "the port for nanomsg even publisher")
}
