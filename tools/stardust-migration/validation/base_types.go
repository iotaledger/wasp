package validation

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"

	"github.com/ethereum/go-ethereum/core/types"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	old_emulator "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/emulator"
)

const (
	newSchema = allmigrations.SchemaVersionMigratedRebased
)

func oldAgentIDToStr(agentID old_isc.AgentID) string {
	switch agentID := agentID.(type) {
	case *old_isc.AddressAgentID:
		return fmt.Sprintf("AddressAgentID(%v)", agentID.Address().String())
	case *old_isc.ContractAgentID:
		return fmt.Sprintf("ContractAgentID(%v)", agentID.Hname())
	case *old_isc.EthereumAddressAgentID:
		return fmt.Sprintf("EthereumAddressAgentID(%v)", agentID.EthAddress().String())
	case *old_isc.NilAgentID:
		panic(fmt.Sprintf("Found agent ID with kind = AgentIDIsNil: %v", agentID))
	default:
		panic(fmt.Sprintf("Unknown agent ID kind: %v/%T = %v", agentID.Kind(), agentID, agentID))
	}
}

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

func oldBlockToStr(block *old_blocklog.BlockInfo) string {
	anchorPresent := block.PreviousAliasOutput != nil &&
		block.PreviousAliasOutput.GetAliasOutput() != nil

	return fmt.Sprintf("v=%v, i=%v, t=%v, tr=%v, sr=%v, or=%v, an=%v, gb=%v, gfc=%v",
		allmigrations.SchemaVersionMigratedRebased, // Version is not migrated, so we just set here expected new value
		block.BlockIndex(),
		block.Timestamp.Unix(),
		block.TotalRequests,
		block.NumSuccessfulRequests,
		block.NumOffLedgerRequests,
		lo.Ternary(anchorPresent, "present", "missing"),
		block.GasBurned,
		block.GasFeeCharged,
	)
}

func newBlockToStr(block *blocklog.BlockInfo) string {
	// TODO: We do not validate actual values of the anchor here. Should we?
	anchorPresent := block.PreviousAnchor != nil &&
		block.PreviousAnchor.Anchor() != nil &&
		block.PreviousAnchor.Anchor().ObjectID != nil &&
		block.PreviousAnchor.Anchor().Digest != nil

	if anchorPresent {
		if block.PreviousAnchor.GetStateIndex() != block.BlockIndex-1 {
			panic(fmt.Sprintf("invalid previoud anchor in the block: expected anchor for block %v, got for %v",
				block.BlockIndex-1, block.PreviousAnchor.GetStateIndex()))
		}
	}

	return fmt.Sprintf("v=%v, i=%v, t=%v, tr=%v, sr=%v, or=%v, an=%v, gb=%v, gfc=%v",
		block.SchemaVersion,
		block.BlockIndex,
		block.Timestamp.Unix(),
		block.TotalRequests,
		block.NumSuccessfulRequests,
		block.NumOffLedgerRequests,
		lo.Ternary(anchorPresent, "present", "missing"),
		// Not adding L1 params here although they are part of new block.
		// TODO: Should we add it here?
		block.GasBurned,
		convertNewBaseBalanceToOldBaseBalance(block.GasFeeCharged), // reverse conversion
	)
}

func convertNewBaseBalanceToOldBaseBalance(v coin.Value) uint64 {
	return uint64(v) / 1000
}

func oldEVMBlockHeaderToStr(header *old_emulator.Header) string {
	return fmt.Sprintf("h=%v, gl=%v, gu=%v, t=%v, th=%v, rh=%v, b=%x",
		header.Hash.String(),
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.TxHash.String(),
		header.ReceiptHash.String(),
		lo.Must(header.Bloom.MarshalText()),
	)
}

func evmReceiptsToStr(r *types.Receipt) string {
	return string(lo.Must(r.MarshalJSON()))
}
