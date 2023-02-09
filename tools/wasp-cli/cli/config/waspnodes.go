package config

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

const defaultWasp = "defaultWasp"

func GetDefaultWaspNode() string {
	defaultWaspNode := viper.GetString(defaultWasp)
	if defaultWaspNode == "" {
		log.Fatalf("No default wasp node set. Call `nodes add <name> <api> --default` or `set %s <name>`", defaultWasp)
	}
	return defaultWaspNode
}

func SetDefaultWaspNode(nodeName string) {
	Set(defaultWasp, nodeName)
}

func AddWaspNode(nodeName, api string) {
	Set(fmt.Sprintf("wasp.%s.api", nodeName), api)
	if GetDefaultWaspNode() == "" {
		SetDefaultWaspNode(nodeName)
	}
}
