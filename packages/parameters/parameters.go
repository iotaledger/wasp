package parameters

import (
	"github.com/iotaledger/wasp/plugins/config"
	flag "github.com/spf13/pflag"
)

const (
	LoggerLevel             = "logger.level"
	LoggerDisableCaller     = "logger.disableCaller"
	LoggerDisableStacktrace = "logger.disableStacktrace"
	LoggerEncoding          = "logger.encoding"
	LoggerOutputPaths       = "logger.outputPaths"
	LoggerDisableEvents     = "logger.disableEvents"

	DatabaseDir      = "database.directory"
	DatabaseInMemory = "database.inMemory"

	WebAPIBindAddress = "webapi.bindAddress"

	VMBinaryDir     = "vm.binaries"
	VMDefaultVmType = "vm.defaultvm"

	NodeAddress = "nodeconn.address"

	PeeringMyNetId = "peering.netid"
	PeeringPort    = "peering.port"

	NanomsgPublisherPort = "nanomsg.port"
)

func InitFlags() {
	flag.String(LoggerLevel, "info", "log level")
	flag.Bool(LoggerDisableCaller, false, "disable caller info in log")
	flag.Bool(LoggerDisableStacktrace, false, "disable stack trace in log")
	flag.String(LoggerEncoding, "console", "log encoding")
	flag.StringSlice(LoggerOutputPaths, []string{"stdout", "goshimmer.log"}, "log output paths")
	flag.Bool(LoggerDisableEvents, true, "disable logger events")

	flag.String(DatabaseDir, "waspdb", "path to the database folder")
	flag.Bool(DatabaseInMemory, false, "whether the database is only kept in memory and not persisted")

	flag.String(WebAPIBindAddress, "127.0.0.1:8080", "the bind address for the web API")

	flag.String(VMBinaryDir, "wasm", "path where Wasm binaries are located (using file:// schema")
	flag.String(VMDefaultVmType, "dummmy", "default VM type")

	flag.String(NodeAddress, "127.0.0.1:5000", "node host address")

	flag.Int(PeeringPort, 4000, "port for Wasp committee connection/peering")
	flag.String(PeeringMyNetId, "127.0.0.1:4000", "node host address as it is recognized by other peers")

	flag.Int(NanomsgPublisherPort, 5550, "the port for nanomsg even publisher")
}

func GetBool(name string) bool {
	return config.Node.GetBool(name)
}

func GetString(name string) string {
	return config.Node.GetString(name)
}

func GetInt(name string) int {
	return config.Node.GetInt(name)
}
