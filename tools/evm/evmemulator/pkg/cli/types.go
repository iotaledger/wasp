package cli

var (
	ListenAddress       string
	EngineListenAddress string
	GenesisJsonPath     string
	NodeLaunchMode      string
	RemoteHost          string
	LogBodies           bool
	IsHive              bool
)

type TNodeLaunchMode string

const (
	EnumNodeLaunchModeStandalone    TNodeLaunchMode = "standalone"
	EnumNodeLaunchModeDockerCompose TNodeLaunchMode = "docker-compose"
)
