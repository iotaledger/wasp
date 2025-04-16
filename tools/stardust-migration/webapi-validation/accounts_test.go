package webapi_validation

import (
	"fmt"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/samber/lo"
)

const (
	IotaAddress = "0x00cb4d8be2289e430c14dde1a94b1478e27197ce301414bdb29126cf513cb191"
	EvmAdress   = "0x484224e1299ab9b83Cd42ffbd11EAe36A39753B5"
	ChainID     = base.MainnetChainID
)

func Test_OldIotaAgentIDParsing(t *testing.T) {
	chainID := lo.Must(isc.ChainIDFromString("0xdfe1d481a018cfc1a3f64412e6e0ef71d3146b5864ffa5f2e0173d993bbf805e"))

	old := OldAgentIDFromString(IotaAddress, chainID)

	fmt.Printf("old: %v\n", old.String())
}

func Test_OldEvmAgentIDParsing(t *testing.T) {
	fmt.Printf("oldAgentID: %v\n", "")
	chainID := lo.Must(isc.ChainIDFromString("0xdfe1d481a018cfc1a3f64412e6e0ef71d3146b5864ffa5f2e0173d993bbf805e"))

	old := OldAgentIDFromString(EvmAdress, chainID)

	fmt.Printf("old: %v\n", old.String())
}

func Test_NewIotaAgentIDParsing(t *testing.T) {
	new, err := isc.AgentIDFromString(IotaAddress)
	if err != nil {
		panic(err)
	}
	fmt.Printf("new: %v\n", new.String())
}

func Test_NewEvmAgentIDParsing(t *testing.T) {
	new, err := isc.AgentIDFromString(EvmAdress)
	if err != nil {
		panic(err)
	}
	fmt.Printf("new: %v\n", new.String())
}

func OldAgentIDFromString(s string, chainID isc.ChainID) old_isc.AgentID {

	// allow EVM addresses as AgentIDs without the chain specified
	if strings.HasPrefix(s, "0x") && !strings.Contains(s, old_isc.AgentIDStringSeparator) {
		s = s + old_isc.AgentIDStringSeparator + chainID.String()
	}
	agentID, err := old_isc.AgentIDFromString(s)
	if err != nil {
		panic(err)
	}
	return agentID
}

// 2025/04/15 13:25:55 accKey = 0100f575115c60bd20111f4d29d0a6379c1ddc402d358985906cb3b2b53d1cf5208e / �u\`� M)Ц7��@-5���l���=� �,
// accID = iota1qr6h2y2uvz7jqyglf55apf3hnswacspdxkyctyrvkwet20gu75sguf5p97l,
// accStr = AddressAgentID(0xf575115c60bd20111f4d29d0a6379c1ddc402d358985906cb3b2b53d1cf5208e)
func newAgentIDToStr(agentID isc.AgentID) string {
	switch agentID := agentID.(type) {
	case *isc.AddressAgentID:
		return fmt.Sprintf("AddressAgentID(%v)", agentID.Address().String())
	case *isc.ContractAgentID:
		return fmt.Sprintf("ContractAgentID(%v)", agentID.Hname())
	case *isc.EthereumAddressAgentID:
		return fmt.Sprintf("EthereumAddressAgentID(%v)", agentID.EthAddress().String())
	case *isc.NilAgentID:
		panic(fmt.Sprintf("Found agent ID with kind = AgentIDIsNil: %v", agentID))
	default:
		panic(fmt.Sprintf("Unknown agent ID kind: %v/%T = %v", agentID.Kind(), agentID, agentID))
	}
}
