//go:build !noevm
// +build !noevm

package processors

import "github.com/iotaledger/wasp/contracts/native/evm/evmlight"

func init() {
	nativeContracts = append(nativeContracts, evmlight.Processor)
}
