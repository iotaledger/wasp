package cli

var ListenAddress string
var GenesisJsonPath string
var NodeLaunchMode string

type TNodeLaunchMode string

const (
	EnumNodeLaunchModeStandalone    TNodeLaunchMode = "standalone"
	EnumNodeLaunchModeDockerCompose TNodeLaunchMode = "docker-compose"
)
