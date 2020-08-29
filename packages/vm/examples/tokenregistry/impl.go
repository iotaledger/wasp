// smart contract implements Token Registry. User can mint any number of new colored tokens to own address
// and in the same transaction can register the whole Supply of new tokens in the TokenRegistry.
// TokenRegistry contains metadata. It can be changed by the owner of the record
// Initially the owner is the minter. Owner can transfer ownership of the metadata record to another address
package tokenregistry

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "8h2RGcbsUgKckh9rZ4VUF75NUfxP4bj1FC66oSF9us6p"

const (
	RequestInitSC            = sctransaction.RequestCode(uint16(0)) // NOP
	RequestMintSupply        = sctransaction.RequestCode(uint16(1))
	RequestUpdateMetadata    = sctransaction.RequestCode(uint16(2))
	RequestTransferOwnership = sctransaction.RequestCode(uint16(3))

	// state vars
	VarStateTheRegistry = "tr"
	VarStateListColors  = "lc" // for testing only

	// request vars
	VarReqDescription         = "dscr"
	VarReqUserDefinedMetadata = "ud"
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

type TokenMetadata struct {
	Supply      int64
	MintedBy    address.Address // originator
	Owner       address.Address // who can update metadata
	Created     int64           // when created record
	Updated     int64           // when last updated
	Description string          // any text
	UserDefined []byte          // any other data (marshalled json etc)
}

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (v tokenRegistryProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

func (v tokenRegistryProcessor) GetDescription() string {
	return "TokenRegistry hard coded smart contract processor"
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
	ctx.Publish("TokenRegistry: mintSupply")

	reqAccess := ctx.AccessRequest()
	reqId := reqAccess.ID()
	colorOfTheSupply := (balance.Color)(*reqId.TransactionId())

	registry := ctx.AccessState().GetDictionary(VarStateTheRegistry)
	if registry.GetAt(colorOfTheSupply[:]) != nil {
		// already exist
		ctx.Publishf("TokenRegistry: Supply of color %s already exist", colorOfTheSupply.String())
		return
	}
	supply := reqAccess.NumFreeMintedTokens()
	if supply <= 0 {
		// no tokens were minted on top of request tokens
		ctx.Publish("TokenRegistry: the free minted Supply must be > 0")
		return

	}
	description, ok, err := reqAccess.Args().GetString(VarReqDescription)
	if err != nil {
		ctx.Publish("TokenRegistry: inconsistency 1")
		return
	}
	if !ok {
		description = "no dscr"
	}
	uddata, err := reqAccess.Args().Get(VarReqUserDefinedMetadata)
	if err != nil {
		ctx.Publish("TokenRegistry: inconsistency 2")
		return
	}
	rec := &TokenMetadata{
		Supply:      supply,
		MintedBy:    reqAccess.Sender(),
		Owner:       reqAccess.Sender(),
		Created:     ctx.GetTimestamp(),
		Updated:     ctx.GetTimestamp(),
		Description: description,
		UserDefined: uddata,
	}
	data, err := util.Bytes(rec)
	if err != nil {
		ctx.Publish("TokenRegistry: inconsistency 3")
		return
	}
	registry.SetAt(colorOfTheSupply[:], data)

	stateAccess := ctx.AccessState()
	lst, ok := stateAccess.GetString(VarStateListColors)
	if !ok {
		lst = colorOfTheSupply.String()
	} else {
		lst += ", " + colorOfTheSupply.String()
	}
	stateAccess.SetString(VarStateListColors, lst)

	ctx.Publishf("TokenRegistry.mintSupply: success. Color: %s, Owner: %s, Description: '%s' User defined data: '%s'",
		colorOfTheSupply.String(), rec.Owner.String(), rec.Description, string(rec.UserDefined))
}

func updateMetadata(ctx vmtypes.Sandbox) {
	ctx.Publishf("TokenRegistry: updateMetadata not implemented")
}

func transferOwnership(ctx vmtypes.Sandbox) {
	ctx.Publishf("TokenRegistry: transferOwnership not implemented")
}
