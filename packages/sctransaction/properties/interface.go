package properties

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
)

func (prop *properties) SenderAddress() *address.Address {
	return &prop.senderAddress
}

func (prop *properties) IsState() bool {
	return prop.isState
}

func (prop *properties) IsOrigin() bool {
	return prop.isState
}

func (prop *properties) MustChainID() *coretypes.ChainID {
	if !prop.isState {
		panic("MustChainID: must be a state transaction")
	}
	return &prop.chainID
}

func (prop *properties) MustStateColor() *balance.Color {
	if !prop.isState {
		panic("MustStateColor: must be a state transaction")
	}
	return &prop.stateColor
}

// NumFreeMintedTokens return total minted tokens minus number of requests
// all of those tokens will be minted to other addresses than chain address
// after all requests will be settled, the remaining minted tokens will be free minted tokens.
func (prop *properties) NumFreeMintedTokens() int64 {
	if prop.isOrigin {
		return 0
	}
	return prop.numTotalMintedTokens - int64(prop.numRequests)
}

func (prop *properties) FreeTokensForAddress(addr address.Address) coretypes.ColoredBalances {
	if ret, ok := prop.freeTokensByAddress[addr]; ok {
		return ret
	}
	return cbalances.Nil
}

func (prop *properties) String() string {
	ret := "---- Transaction:\n"
	ret += fmt.Sprintf("   requests: %d\n", prop.numRequests)
	ret += fmt.Sprintf("   senderAddress: %s\n", prop.senderAddress.String())
	ret += fmt.Sprintf("   isState: %v\n   isOrigin: %v\n", prop.isState, prop.isOrigin)
	ret += fmt.Sprintf("   chainAddress: %s\n", prop.chainAddress.String())
	ret += fmt.Sprintf("   chainID: %s\n   stateColor: %s\n", prop.chainID.String(), prop.stateColor.String())
	ret += fmt.Sprintf("   timestamp: %d\n    stateHash: %s\n", prop.timestamp, prop.stateHash.String())
	ret += fmt.Sprintf("   numMinted: %d\n", prop.numTotalMintedTokens)
	ret += fmt.Sprintf("   data payload size: %d\n", prop.dataPayloadSize)
	return ret
}
