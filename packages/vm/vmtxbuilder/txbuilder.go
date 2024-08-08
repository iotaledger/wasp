package vmtxbuilder

import (
	"math/big"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// AnchorTransactionBuilder represents structure which handles all the data needed to eventually
// build an essence of the anchor transaction
type AnchorTransactionBuilder struct {
	iscPackage sui.Address

	// anchorOutput output of the chain
	anchor *iscmove.RefWithObject[iscmove.Anchor]

	// already consumed requests, specified by entire Request. It is needed for checking validity
	consumed []isc.OnLedgerRequest
}

var _ TransactionBuilder = &AnchorTransactionBuilder{}

// NewAnchorTransactionBuilder creates new AnchorTransactionBuilder object
func NewAnchorTransactionBuilder(
	iscPackage sui.Address,
	anchor *iscmove.RefWithObject[iscmove.Anchor],
) *AnchorTransactionBuilder {
	return &AnchorTransactionBuilder{
		iscPackage: iscPackage,
		anchor:     anchor,
	}
}

func (txb *AnchorTransactionBuilder) Clone() TransactionBuilder {
	a := *txb.anchor
	newConsumed := make([]isc.OnLedgerRequest, len(txb.consumed))
	for i, v := range txb.consumed {
		newConsumed[i] = v.Clone()
	}
	return &AnchorTransactionBuilder{
		anchor:   &a,
		consumed: newConsumed,
	}
}

// ConsumeRequest adds an input to the transaction.
// It panics if transaction cannot hold that many inputs
// All explicitly consumed inputs will hold fixed index in the transaction
// Returns  the amount of baseTokens needed to cover SD costs for the NTs/NFT contained by the request output
func (txb *AnchorTransactionBuilder) ConsumeRequest(req isc.OnLedgerRequest) {
	// TODO we may need to assert the maximum size of the request we can consume here

	txb.consumed = append(txb.consumed, req)
}

func (txb *AnchorTransactionBuilder) SendObject(object sui.Object) (storageDepositReturned *big.Int) {
	return nil
}
func (txb *AnchorTransactionBuilder) BuildTransactionEssence(stateRoot *state.L1Commitment) sui.ProgrammableTransaction {
	ptb, err := iscmoveclient.NewReceiveRequestPTB(
		txb.iscPackage,
		&txb.anchor.ObjectRef,
		onRequestsToRequestRefs(txb.consumed),
		onRequestsToAssetsBagMap(txb.consumed),
		stateRoot.TrieRoot().Bytes(),
	)
	if err != nil {
		panic(err)
	}
	return ptb
}

func onRequestsToRequestRefs(reqs []isc.OnLedgerRequest) []sui.ObjectRef {
	refs := make([]sui.ObjectRef, len(reqs))
	for i, req := range reqs {
		refs[i] = req.RequestRef()
	}
	return refs
}

func onRequestsToAssetsBagMap(reqs []isc.OnLedgerRequest) map[sui.ObjectRef]*iscmove.AssetsBagWithBalances {
	m := make(map[sui.ObjectRef]*iscmove.AssetsBagWithBalances)
	for _, req := range reqs {
		assetsBagWithBalances := &iscmove.AssetsBagWithBalances{
			AssetsBag: *req.AssetsBag(),
			Balances:  make(iscmove.AssetsBagBalances),
		}
		assets := req.Assets()
		for k, v := range assets.Coins {
			panic("refactor me: Change BigInt to coin.Value (aka uint64)")

			assetsBagWithBalances.Balances[suijsonrpc.CoinType(k)] = &suijsonrpc.Balance{
				CoinType:     suijsonrpc.CoinType(k),
				TotalBalance: &suijsonrpc.BigInt{big.NewInt(0).SetUint64(uint64(v))},
			}

		}
		m[req.RequestRef()] = assetsBagWithBalances

	}
	return m
}
