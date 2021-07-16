package core

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
)

func init() {
	// printWellKnownHnames()
}

// for debugging
func PrintWellKnownHnames() {
	fmt.Printf("--------------- well known hnames ------------------\n")
	hashes := make([]hashing.HashValue, 0)
	for _, rec := range AllCoreContractsByHash {
		hashes = append(hashes, rec.Contract.ProgramHash)
	}
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i][:], hashes[j][:]) < 0
	})
	for _, h := range hashes {
		rec := AllCoreContractsByHash[h]
		fmt.Printf("    %10d, %10s: '%s'\n", rec.Contract.Hname(), rec.Contract.Hname(), rec.Contract.Name)
	}
	fmt.Printf("    %10d, %10s: '%s'\n", iscp.EntryPointInit, iscp.EntryPointInit, iscp.FuncInit)
	fmt.Printf("    %10d, %10s: '%s'\n", iscp.Hn("testcore"), iscp.Hn("testcore"), "testcore")
	fmt.Printf("--------------- well known hnames ------------------\n")
}
