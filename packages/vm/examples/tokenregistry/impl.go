// smart contract code implements Token Registry. User can mint any number of new colored tokens to own address
// and in the same transaction can register the whole Supply of new tokens in the TokenRegistry.
// TokenRegistry contains metadata of the supply minted this way. It can be changed by the owner of the record
// Initially the owner is the minter. Owner can transfer ownership of the metadata record to another address
package tokenregistry

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// program hash is an ID of the smart contract program
const ProgramHash = "8h2RGcbsUgKckh9rZ4VUF75NUfxP4bj1FC66oSF9us6p"
const Description = "TokenRegistry, a PoC smart contract"

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

// implement Processor and EntryPoint interfaces

type tokenRegistryProcessor map[sctransaction.RequestCode]tokenRegistryEntryPoint

type tokenRegistryEntryPoint func(ctx vmtypes.Sandbox)

// the processor is a map of entry points
var entryPoints = tokenRegistryProcessor{
	RequestInitSC:            initSC,
	RequestMintSupply:        mintSupply,
	RequestUpdateMetadata:    updateMetadata,
	RequestTransferOwnership: transferOwnership,
}

// TokenMetadata is a structure for one supply
type TokenMetadata struct {
	Supply      int64
	MintedBy    address.Address // originator
	Owner       address.Address // who can update metadata
	Created     int64           // when created record
	Updated     int64           // when recordt last updated
	Description string          // any text
	UserDefined []byte          // any other data (marshalled json etc)
}

// Point to link statically with the Wasp
func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (v tokenRegistryProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

func (v tokenRegistryProcessor) GetDescription() string {
	return "TokenRegistry hard-coded smart contract processor"
}

// Run runs the entry point
func (ep tokenRegistryEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

// WithGasLimit not used
func (ep tokenRegistryEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

// initSC NOP
func initSC(ctx vmtypes.Sandbox) {
	ctx.Publishf("TokenRegistry: initSC")
}

const maxDescription = 150

// mintSupply implements 'mint supply' request
func mintSupply(ctx vmtypes.Sandbox) {
	ctx.Publish("TokenRegistry: mintSupply")

	reqAccess := ctx.AccessRequest()
	reqId := reqAccess.ID()
	colorOfTheSupply := (balance.Color)(*reqId.TransactionId())

	registry := ctx.AccessState().GetDictionary(VarStateTheRegistry)
	// check for duplicated colors
	if registry.GetAt(colorOfTheSupply[:]) != nil {
		// already exist
		ctx.Publishf("TokenRegistry: Supply of color %s already exist", colorOfTheSupply.String())
		return
	}
	// get the number of tokens, which are minted by the request transaction - tokens which are used for requests tracking
	supply := reqAccess.NumFreeMintedTokens()
	if supply <= 0 {
		// no tokens were minted on top of request tokens
		ctx.Publish("TokenRegistry: the free minted Supply must be > 0")
		return

	}
	// get the description of the supply form the request argument
	description, ok, err := reqAccess.Args().GetString(VarReqDescription)
	if err != nil {
		ctx.Publish("TokenRegistry: inconsistency 1")
		return
	}
	if !ok {
		description = "no dscr"
	}
	description = util.GentleTruncate(description, maxDescription)

	// get the additional arbitrary deta attached to the supply record
	uddata, err := reqAccess.Args().Get(VarReqUserDefinedMetadata)
	if err != nil {
		ctx.Publish("TokenRegistry: inconsistency 2")
		return
	}
	// create the metadata record and marshal it into binary
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
	// put the metadata into the dictionary of the registry by color
	registry.SetAt(colorOfTheSupply[:], data)

	// maintain the list all colors in the registry (dictionary keys)
	// only used for assertion in tests
	// TODO not finished
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
	// TODO not implemented
	ctx.Publishf("TokenRegistry: updateMetadata not implemented")
}

func transferOwnership(ctx vmtypes.Sandbox) {
	// TODO not implemented
	ctx.Publishf("TokenRegistry: transferOwnership not implemented")
}
