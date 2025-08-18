package cli

var ListenAddress string
var GenesisJsonPath string
var NodeLaunchMode string
var RemoteHost string

type TNodeLaunchMode string

const (
	EnumNodeLaunchModeStandalone    TNodeLaunchMode = "standalone"
	EnumNodeLaunchModeDockerCompose TNodeLaunchMode = "docker-compose"
)
