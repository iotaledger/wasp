// smart contract implements Token Registry. User can mint any number of new colored tokens to own address
// and in the same transaction can register the whole supply of new tokens in the TokenRegistry.
// TokenRegistry contains metadata. It can be changed by the owner of the record
// Initially the owner is the minter. Owner can transfer ownership of the metadata record to another address
package tokenregistry

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "8h2RGcbsUgKckh9rZ4VUF75NUfxP4bj1FC66oSF9us6p"

const (
	RequestInitSC            = sctransaction.RequestCode(uint16(0)) // NOP
	RequestMintSupply        = sctransaction.RequestCode(uint16(1))
	RequestUpdateMetadata    = sctransaction.RequestCode(uint16(2))
	RequestTransferOwnership = sctransaction.RequestCode(uint16(3))
)

type tokenRegistryProcessor map[sctransaction.RequestCode]tokenRegistryEntryPoint

type tokenRegistryEntryPoint func(ctx vmtypes.Sandbox)

// the processor is a map of entry points
var entryPoints = tokenRegistryProcessor{
	RequestInitSC:            initSC,
	RequestMintSupply:        mintSupply,
	RequestUpdateMetadata:    updateMetadata,
	RequestTransferOwnership: transferOwnership,
}

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (v tokenRegistryProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

// does nothing, i.e. resulting state update is empty
func (ep tokenRegistryEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

func (ep tokenRegistryEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

func initSC(ctx vmtypes.Sandbox) {
	ctx.Publishf("TokenRegistry: initSC")
}

func mintSupply(ctx vmtypes.Sandbox) {
	ctx.Publishf("TokenRegistry: mintSupply")
}

func updateMetadata(ctx vmtypes.Sandbox) {
	ctx.Publishf("TokenRegistry: updateMetadata")
}

func transferOwnership(ctx vmtypes.Sandbox) {
	ctx.Publishf("TokenRegistry: transferOwnership")
}
