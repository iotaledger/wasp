package chain

import "github.com/iotaledger/wasp/tools/wasp-cli/log"

func activateCmd(args []string) {
	log.Check(MultiClient().ActivateChain(GetCurrentChainID()))
}

func deactivateCmd(args []string) {
	log.Check(MultiClient().DeactivateChain(GetCurrentChainID()))
}
