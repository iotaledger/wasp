package peering

import (
	flag "github.com/spf13/pflag"
)

func init() {
	flag.Int(CfgPeeringPort, 4000, "port for Wasp committee connection/peering")
	flag.String(CfgMyNetId, "127.0.0.1:4000", "node host address as it is recognized by other peers")
}

const (
	CfgMyNetId     = "peering.netid"
	CfgPeeringPort = "peering.port"
)
