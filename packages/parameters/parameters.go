package parameters

import (
	"reflect"
	"unsafe"

	"github.com/iotaledger/hive.go/configuration"
	"github.com/knadh/koanf"
	flag "github.com/spf13/pflag"
)

var all *configuration.Configuration

const (
	LoggerLevel             = "logger.level"
	LoggerDisableCaller     = "logger.disableCaller"
	LoggerDisableStacktrace = "logger.disableStacktrace"
	LoggerEncoding          = "logger.encoding"
	LoggerOutputPaths       = "logger.outputPaths"
	LoggerDisableEvents     = "logger.disableEvents"

	DatabaseDir      = "database.directory"
	DatabaseInMemory = "database.inMemory"

	WebAPIBindAddress    = "webapi.bindAddress"
	WebAPIAdminWhitelist = "webapi.adminWhitelist"
	WebAPIAuth           = "webapi.auth"

	DashboardBindAddress       = "dashboard.bindAddress"
	DashboardExploreAddressURL = "dashboard.exploreAddressUrl"
	DashboardAuth              = "dashboard.auth"

	NodeAddress = "nodeconn.address"

	PeeringMyNetID                   = "peering.netid"
	PeeringPort                      = "peering.port"
	PeeringNeighbors                 = "peering.neighbors"
	PullMissingRequestsFromCommittee = "peering.pullMissingRequests"

	NanomsgPublisherPort = "nanomsg.port"

	IpfsGatewayAddress = "ipfs.gatewayAddress"

	OffledgerBroadcastUpToNPeers = "offledger.broadcastUpToNPeers"
	OffledgerBroadcastInterval   = "offledger.broadcastInterval"
	OffledgerAPICacheTTL         = "offledger.apiCacheTTL"

	ProfilingBindAddress   = "profiling.bindAddress"
	ProfilingEnabled       = "profiling.enabled"
	ProfilingWriteProfiles = "profiling.writeProfiles"

	PrometheusBindAddress = "prometheus.bindAddress"
	PrometheusEnabled     = "prometheus.enabled"
)

func Init() *configuration.Configuration {
	// set the default logger config
	all = configuration.New()

	flag.String(LoggerLevel, "info", "log level")
	flag.Bool(LoggerDisableCaller, false, "disable caller info in log")
	flag.Bool(LoggerDisableStacktrace, false, "disable stack trace in log")
	flag.String(LoggerEncoding, "console", "log encoding")
	flag.StringSlice(LoggerOutputPaths, []string{"stdout", "goshimmer.log"}, "log output paths")
	flag.Bool(LoggerDisableEvents, true, "disable logger events")

	flag.String(DatabaseDir, "waspdb", "path to the database folder")
	flag.Bool(DatabaseInMemory, false, "whether the database is only kept in memory and not persisted")

	flag.String(WebAPIBindAddress, "127.0.0.1:8080", "the bind address for the web API")
	flag.StringSlice(WebAPIAdminWhitelist, []string{}, "IP whitelist for /adm wndpoints")
	flag.StringToString(WebAPIAuth, nil, "authentication scheme for web API")

	flag.String(DashboardBindAddress, "127.0.0.1:7000", "the bind address for the node dashboard")
	flag.String(DashboardExploreAddressURL, "", "URL to add as href to addresses in the dashboard [default: <nodeconn.address>:8081/explorer/address]")
	flag.StringToString(DashboardAuth, nil, "authentication scheme for the node dashboard")

	flag.String(NodeAddress, "127.0.0.1:5000", "node host address")

	flag.Int(PeeringPort, 4000, "port for Wasp committee connection/peering")
	flag.String(PeeringMyNetID, "127.0.0.1:4000", "node host address as it is recognized by other peers")
	flag.StringSlice(PeeringNeighbors, []string{}, "list of neighbors: known peer netIDs")
	flag.Bool(PullMissingRequestsFromCommittee, true, "whether or not to pull missing requests from other committee members")

	flag.Int(NanomsgPublisherPort, 5550, "the port for nanomsg even publisher")

	flag.String(IpfsGatewayAddress, "https://ipfs.io/", "the address of HTTP(s) gateway to which download from ipfs requests will be forwarded")

	flag.Int(OffledgerBroadcastUpToNPeers, 2, "number of peers an offledger request is broadcasted to")
	flag.Int(OffledgerBroadcastInterval, 5000, "time between re-broadcast of offledger requests (in ms)")
	flag.Int(OffledgerAPICacheTTL, 5*60, "time to keep processed offledger requests in api cache (in seconds)")

	flag.String(ProfilingBindAddress, "127.0.0.1:6060", "pprof http server address")
	flag.Bool(ProfilingEnabled, false, "whether profiling is enabled")
	flag.Bool(ProfilingWriteProfiles, false, "whether to write profiling profiles to disk on node shutdown (when enabled some metrics will be unavailable via pprof runtime endpoint)")

	flag.String(PrometheusBindAddress, "127.0.0.1:2112", "prometheus metrics http server address")
	flag.Bool(PrometheusEnabled, false, "disable and enable prometheus metrics")

	return all
}

func GetBool(name string) bool {
	return all.Bool(name)
}

func GetString(name string) string {
	return all.String(name)
}

func GetStringSlice(name string) []string {
	return all.Strings(name)
}

func GetInt(name string) int {
	return all.Int(name)
}

func GetStringToString(name string) map[string]string {
	return all.StringMap(name)
}

func Dump() map[string]interface{} {
	// hack to access private member Node.config
	rf := reflect.ValueOf(all).Elem().FieldByName("config")
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	tree := rf.Interface().(*koanf.Koanf).Raw()

	m := map[string]interface{}{}
	flatten(m, tree, "")
	return m
}

func flatten(dst, src map[string]interface{}, path string) {
	for k, v := range src {
		switch vt := v.(type) {
		case map[string]interface{}:
			flatten(dst, vt, path+k+".")
		default:
			dst[path+k] = v
		}
	}
}
