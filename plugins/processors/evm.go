// +build evm

package processors

import "github.com/iotaledger/wasp/contracts/native/evmchain"

func init() {
	nativeContracts = append(nativeContracts, evmchain.Interface)
}
