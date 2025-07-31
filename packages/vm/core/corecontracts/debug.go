// Package corecontracts defines and manages the core smart contracts that are built into the
// IOTA Wasp node. These contracts provide essential system functionality such as chain governance,
// account management, block logging, error handling, and EVM compatibility.
// The package maintains a registry of core contracts and provides utilities for accessing
// and verifying these contracts across the IOTA Smart Contract platform.
package corecontracts

import (
	"fmt"
	"sort"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

func init() {
	// printWellKnownHnames()
}

// PrintWellKnownHnames prints all well-known contract hnames and their corresponding names in a sorted manner.
func PrintWellKnownHnames() {
	fmt.Printf("--------------- well known hnames ------------------\n")
	hnames := make([]isc.Hname, 0)
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
	fmt.Printf("    %10d, %10s: '%s'\n", isc.Hn("testcore"), isc.Hn("testcore"), "testcore")
	fmt.Printf("--------------- well known hnames ------------------\n")
}
