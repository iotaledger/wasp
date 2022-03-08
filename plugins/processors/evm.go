//go:build !noevm
// +build !noevm

package processors

import "github.com/iotaledger/wasp/contracts/native/evm/evmimpl"

func init() {
	nativeContracts = append(
		nativeContracts,
		evmimpl.Processor,
	)
}
