package corecontracts

import (
	"fmt"
	"sort"

	"github.com/iotaledger/wasp/packages/iscp"
)

func init() {
	// printWellKnownHnames()
}

// for debugging
func PrintWellKnownHnames() {
	fmt.Printf("--------------- well known hnames ------------------\n")
	hnames := make([]iscp.Hname, 0)
	for h := range All {
		hnames = append(hnames, h)
	}
	sort.Slice(hnames, func(i, j int) bool {
		return hnames[i] < hnames[j]
	})
	for _, h := range hnames {
		rec := All[h]
		fmt.Printf("    %10d, %10s: '%s'\n", rec.Hname(), rec.Hname(), rec.Name)
	}
	fmt.Printf("    %10d, %10s: '%s'\n", iscp.EntryPointInit, iscp.EntryPointInit, iscp.FuncInit)
	fmt.Printf("    %10d, %10s: '%s'\n", iscp.Hn("testcore"), iscp.Hn("testcore"), "testcore")
	fmt.Printf("--------------- well known hnames ------------------\n")
}
