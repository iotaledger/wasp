package cli

var ListenAddress string
var EngineListenAddress string
var GenesisJsonPath string
var NodeLaunchMode string
var RemoteHost string
var LogBodies bool
var IsHive bool

type TNodeLaunchMode string

const (
	EnumNodeLaunchModeStandalone    TNodeLaunchMode = "standalone"
	EnumNodeLaunchModeDockerCompose TNodeLaunchMode = "docker-compose"
)
