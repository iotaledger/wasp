package peering

import (
	"flag"
)

const CfgPeeringPort = "peering.port"

func init() {
	// TODO default doesn't work
	flag.Int(CfgPeeringPort, 4000, "port for Wasp committee connection/peering")
}
