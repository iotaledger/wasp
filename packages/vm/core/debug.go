package core

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func init() {
	//printWellKnownHnames()
}

// for debugging
func printWellKnownHnames() {
	fmt.Printf("--------------- well known hnames ------------------\n")
	fmt.Printf("    %10s: '%s'\n", root.Interface.Hname().String(), root.Interface.Name)
	fmt.Printf("    %10s: '%s'\n", accounts.Interface.Hname().String(), accounts.Interface.Name)
	fmt.Printf("    %10s: '%s'\n", blob.Interface.Hname().String(), blob.Interface.Name)
	fmt.Printf("    %10s: '%s'\n", eventlog.Interface.Hname().String(), eventlog.Interface.Name)
	fmt.Printf("    %10s: '%s'\n", coretypes.EntryPointInit.String(), coretypes.FuncInit)
	fmt.Printf("--------------- well known hnames ------------------\n")
}
