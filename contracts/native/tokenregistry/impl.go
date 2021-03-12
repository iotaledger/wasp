// +build ignore

// smart contract code implements Token Registry. User can mint any number of new colored tokens to own address
// and in the same transaction can register the whole Supply of new tokens in the TokenRegistry.
// TokenRegistry contains metadata of the supply minted this way. It can be changed by the owner of the record
// Initially the owner is the minter. Owner can transfer ownership of the metadata record to another address
package tokenregistry

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

// program hash is an ID of the smart contract program
const ProgramHash = "8h2RGcbsUgKckh9rZ4VUF75NUfxP4bj1FC66oSF9us6p"
const Description = "TokenRegistry, a PoC smart contract"

var (
	RequestMintSupply        = coretypes.Hn("mintSupply")
	RequestUpdateMetadata    = coretypes.Hn("updateMetadata")
	RequestTransferOwnership = coretypes.Hn("transferOwnership")
)

const (

	// state vars
	VarStateTheRegistry = "tr"
	VarStateListColors  = "lc" // for testing only

	// request vars
	VarReqDescription         = "dscr"
	VarReqUserDefinedMetadata = "ud"
)

// implement Processor and EntryPoint interfaces

type tokenRegistryProcessor map[coretypes.Hname]tokenRegistryEntryPoint

type tokenRegistryEntryPoint func(ctx coretypes.Sandbox) error

// the processor is a map of entry points
var entryPoints = tokenRegistryProcessor{
	RequestMintSupply:        mintSupply,
	RequestUpdateMetadata:    updateMetadata,
	RequestTransferOwnership: transferOwnership,
}

// TokenMetadata is a structure for one supply
type TokenMetadata struct {
	Supply      int64
	MintedBy    coretypes.AgentID // originator
	Owner       coretypes.AgentID // who can update metadata
	Created     int64             // when created record
	Updated     int64             // when recordt last updated
	Description string            // any text
	UserDefined []byte            // any other data (marshalled json etc)
}

// Point to link statically with the Wasp
func GetProcessor() coretypes.Processor {
	return entryPoints
}

func (v tokenRegistryProcessor) GetEntryPoint(code coretypes.Hname) (coretypes.EntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

func (v tokenRegistryProcessor) GetDescription() string {
	return "TokenRegistry hard-coded smart contract processor"
}

// Run runs the entry point
func (ep tokenRegistryEntryPoint) Call(ctx coretypes.Sandbox) (dict.Dict, error) {
	err := ep(ctx)
	if err != nil {
		ctx.Event(fmt.Sprintf("error %v", err))
	}
	return nil, err
}

// TODO
func (ep tokenRegistryEntryPoint) IsView() bool {
	return false
}

// TODO
func (ep tokenRegistryEntryPoint) CallView(ctx coretypes.SandboxView) (dict.Dict, error) {
	panic("implement me")
}

const maxDescription = 150

// mintSupply implements 'mint supply' request
func mintSupply(ctx coretypes.Sandbox) error {
	ctx.Event("TokenRegistry: mintSupply")
	params := ctx.Params()

	reqId := ctx.RequestID()
	colorOfTheSupply := (balance.Color)(*reqId.TransactionID())

	registry := collections.NewMap(ctx.State(), VarStateTheRegistry)
	// check for duplicated colors
	if registry.MustGetAt(colorOfTheSupply[:]) != nil {
		// already exist
		return fmt.Errorf("TokenRegistry: Supply of color %s already exist", colorOfTheSupply.String())
	}
	// get the number of tokens, which are minted by the request transaction - tokens which are used for requests tracking
	//supply := ctx.AccessRequest().NumFreeMintedTokens() TODO
	supply := int64(0) // TODO fake
	if supply <= 0 {
		// no tokens were minted on top of request tokens
		return fmt.Errorf("TokenRegistry: the free minted Supply must be > 0")

	}
	// get the description of the supply form the request argument
	description, ok, err := codec.DecodeString(params.MustGet(VarReqDescription))
	if err != nil {
		return fmt.Errorf("TokenRegistry: inconsistency 1")
	}
	if !ok {
		description = "no dscr"
	}
	description = util.GentleTruncate(description, maxDescription)

	// get the additional arbitrary deta attached to the supply record
	uddata, err := params.Get(VarReqUserDefinedMetadata)
	if err != nil {
		return fmt.Errorf("TokenRegistry: inconsistency 2")
	}
	// create the metadata record and marshal it into binary
	senderAddress := ctx.Caller()
	rec := &TokenMetadata{
		Supply:      supply,
		MintedBy:    senderAddress,
		Owner:       senderAddress,
		Created:     ctx.GetTimestamp(),
		Updated:     ctx.GetTimestamp(),
		Description: description,
		UserDefined: uddata,
	}
	data, err := util.Bytes(rec)
	if err != nil {
		return fmt.Errorf("TokenRegistry: inconsistency 3")
	}
	// put the metadata into the dictionary of the registry by color
	registry.MustSetAt(colorOfTheSupply[:], data)

	// maintain the list all colors in the registry (dictionary keys)
	// only used for assertion in tests
	// TODO not finished
	stateAccess := ctx.State()
	lst, ok, _ := codec.DecodeString(stateAccess.MustGet(VarStateListColors))
	if !ok {
		lst = colorOfTheSupply.String()
	} else {
		lst += ", " + colorOfTheSupply.String()
	}
	stateAccess.Set(VarStateListColors, codec.EncodeString(lst))

	ctx.Event(fmt.Sprintf("TokenRegistry.mintSupply: success. Color: %s, Owner: %s, Description: '%s' User defined data: '%s'",
		colorOfTheSupply.String(), rec.Owner.String(), rec.Description, string(rec.UserDefined)))
	return nil
}

func updateMetadata(ctx coretypes.Sandbox) error {
	// TODO not implemented
	return fmt.Errorf("TokenRegistry: updateMetadata not implemented")
}

func transferOwnership(ctx coretypes.Sandbox) error {
	// TODO not implemented
	return fmt.Errorf("TokenRegistry: transferOwnership not implemented")
}
