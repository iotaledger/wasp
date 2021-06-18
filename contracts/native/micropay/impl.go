package micropay

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

func initialize(_ coretypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

func publicKey(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(ctx.Caller().Address().Type() != ledgerstate.AliasAddressType, "micropay.publicKey: caller must be an address")

	par := kvdecoder.New(ctx.Params(), ctx.Log())

	pubKeyBin := par.MustGetBytes(ParamPublicKey)
	addr, err := ctx.Utils().ED25519().AddressFromPublicKey(pubKeyBin)
	a.RequireNoError(err)
	a.Require(addr.Equals(ctx.Caller().Address()), "public key does not correspond to the caller's address")

	pkRegistry := collections.NewMap(ctx.State(), StateVarPublicKeys)
	a.RequireNoError(pkRegistry.SetAt(addr.Bytes(), pubKeyBin))
	return nil, nil
}

// addWarrant adds payment warrant for specific service address
// Params:
// - ParamServiceAddress ledgerstate.Address
func addWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Require(ctx.Caller().Address().Type() != ledgerstate.AliasAddressType, "micropay.addWarrant: caller must be an address")
	payerAddr := ctx.Caller().Address()

	a.Require(getPublicKey(ctx.State(), payerAddr, a) != nil,
		fmt.Sprintf("unknown public key for address %s", payerAddr))

	serviceAddr := par.MustGetAddress(ParamServiceAddress)
	addWarrant, _ := ctx.IncomingTransfer().Get(ledgerstate.ColorIOTA)
	a.Require(addWarrant >= MinimumWarrantIotas, fmt.Sprintf("warrant must be larger than %d iotas", MinimumWarrantIotas))

	warrant, revoke, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, a)
	a.Require(revoke == 0, fmt.Sprintf("warrant of %s for %s is being revoked", payerAddr, serviceAddr))

	payerInfo := collections.NewMap(ctx.State(), string(payerAddr.Bytes()))
	setWarrant(payerInfo, serviceAddr, warrant+addWarrant)

	// all non-iota token accrue on-chain to the caller
	sendBack := ctx.IncomingTransfer().Map()
	delete(sendBack, ledgerstate.ColorIOTA)
	if len(sendBack) > 0 {
		a.RequireNoError(vmcontext.Accrue(ctx, ctx.Caller(), ledgerstate.NewColoredBalances(sendBack)))
	}

	ctx.Event(fmt.Sprintf("[micropay.addWarrant] %s increased warrant %d -> %d i for %s",
		payerAddr, warrant, warrant+addWarrant, serviceAddr))
	return nil, nil
}

// revokeWarrant revokes payment warrant for specific service address
// It will be in effect next 1 hour, the will be deleted
// Params:
// - ParamServiceAddress ledgerstate.Address
func revokeWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Require(ctx.Caller().Address().Type() != ledgerstate.AliasAddressType, "micropay.addWarrant: caller must be an address")
	payerAddr := ctx.Caller().Address()
	serviceAddr := par.MustGetAddress(ParamServiceAddress)

	w, r, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, a)
	a.Require(w > 0, fmt.Sprintf("warrant of %s to %s does not exist", payerAddr, serviceAddr))
	a.Require(r == 0, fmt.Sprintf("warrant of %s to %s is already being revoked", payerAddr, serviceAddr))

	revokeDeadline := getRevokeDeadline(ctx.GetTimestamp())
	payerInfo := collections.NewMap(ctx.State(), string(payerAddr.Bytes()))
	setWarrantRevoke(payerInfo, serviceAddr, revokeDeadline.Unix())

	// send deterred request to self to revoke the warrant
	iota1 := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 1})
	meta := &coretypes.SendMetadata{
		TargetContract: ctx.Contract(),
		EntryPoint:     coretypes.Hn(FuncCloseWarrant),
		Args: codec.MakeDict(map[string]interface{}{
			ParamPayerAddress:   payerAddr,
			ParamServiceAddress: serviceAddr,
		}),
	}
	opts := coretypes.SendOptions{TimeLock: uint32(revokeDeadline.Unix())}
	succ := ctx.Send(ctx.ChainID().AsAddress(), iota1, meta, opts)
	a.Require(succ, "failed to issue deterred 'close warrant' request")
	return nil, nil
}

// closeWarrant can only be sent from self. It closes the warrant account
// - ParamServiceAddress ledgerstate.Address
// - ParamPayerAddress ledgerstate.Address
func closeWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	myAgentId := coretypes.NewAgentID(ctx.ChainID().AsAddress(), ctx.Contract())
	a.Require(ctx.Caller().Equals(myAgentId), "caller must be self")

	par := kvdecoder.New(ctx.Params(), ctx.Log())
	payerAddr := par.MustGetAddress(ParamPayerAddress)
	serviceAddr := par.MustGetAddress(ParamServiceAddress)
	warrant, _, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, assert.NewAssert(ctx.Log()))
	if warrant > 0 {
		tokens := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: warrant})
		succ := ctx.Send(payerAddr, tokens, nil)
		a.Require(succ, "failed to send %d iotas to address %s", warrant, payerAddr)
	}
	deleteWarrant(ctx.State(), payerAddr, serviceAddr)
	return nil, nil
}

// Params:
// - ParamPayerAddress address.address
// - ParamPayments - array of encoded payments
func settle(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	targetAddr := ctx.Caller().Address()
	a.Require(targetAddr.Type() != ledgerstate.AliasAddressType, "micropay.addWarrant: caller must be an address")

	par := kvdecoder.New(ctx.Params(), ctx.Log())
	payerAddr := par.MustGetAddress(ParamPayerAddress)
	payerPubKeyBin := getPublicKey(ctx.State(), payerAddr, a)
	a.Require(payerPubKeyBin != nil, "public key unknown for %s", payerAddr)

	payments := decodePayments(ctx.Params(), a)
	settledSum, notSettled := processPayments(ctx, payments, payerAddr, targetAddr, payerPubKeyBin)
	ctx.Event(fmt.Sprintf("[micropay.settle] settled %d i, num payments: %d, not settled payments: %d, payer: %s, target %s",
		settledSum, len(payments)-len(notSettled), len(notSettled), payerAddr, targetAddr))
	if len(notSettled) > 0 {
		return nil, fmt.Errorf("number of payments failed to settle: %d", len(notSettled))
	}
	return nil, nil
}

// getWarrantInfo return warrant info for given payer and services addresses
// Params:
// - ParamServiceAddress ledgerstate.Address
// - ParamPayerAddress ledgerstate.Address
// Output:
// - ParamWarrant int64 if == 0 no warrant
// - ParamRevoked int64 is exists, timestamp in Unix nanosec when warrant will be revoked
func getWarrantInfo(ctx coretypes.SandboxView) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	payerAddr := par.MustGetAddress(ParamPayerAddress)
	serviceAddr := par.MustGetAddress(ParamServiceAddress)
	warrant, revoke, lastOrd := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, assert.NewAssert(ctx.Log()))
	ret := dict.New()
	if warrant > 0 {
		ret.Set(ParamWarrant, codec.EncodeUint64(warrant))
	}
	if revoke > 0 {
		ret.Set(ParamRevoked, codec.EncodeUint64(revoke))
	}
	if lastOrd > 0 {
		ret.Set(ParamLastOrd, codec.EncodeUint64(lastOrd))
	}
	return ret, nil
}

//  utility

func getWarrantInfoIntern(state kv.KVStoreReader, payer, service ledgerstate.Address, a assert.Assert) (uint64, uint64, uint64) {
	payerInfo := collections.NewMapReadOnly(state, string(payer.Bytes()))
	warrantBin, err := payerInfo.GetAt(service.Bytes())
	a.RequireNoError(err)
	warrant, exists, err := codec.DecodeUint64(warrantBin)
	a.RequireNoError(err)
	if !exists {
		warrant = 0
	}
	revokeBin, err := payerInfo.GetAt(getRevokeKey(service))
	revoke, exists, err := codec.DecodeUint64(revokeBin)
	if !exists {
		revoke = 0
	}
	lastOrdBin, err := payerInfo.GetAt(getLastOrdKey(service))
	lastOrd, exists, err := codec.DecodeUint64(lastOrdBin)
	if !exists {
		lastOrd = 0
	}
	return warrant, revoke, lastOrd
}

func setWarrant(payerAccount *collections.Map, service ledgerstate.Address, value uint64) {
	payerAccount.MustSetAt(service.Bytes(), codec.EncodeUint64(value))
}

func setWarrantRevoke(payerAccount *collections.Map, service ledgerstate.Address, deadline int64) {
	payerAccount.MustSetAt(getRevokeKey(service), codec.EncodeInt64(deadline))
}

func setLastOrd(payerAccount *collections.Map, service ledgerstate.Address, lastOrd uint64) {
	payerAccount.MustSetAt(getLastOrdKey(service), codec.EncodeUint64(lastOrd))
}

func deleteWarrant(state kv.KVStore, payer, service ledgerstate.Address) {
	payerInfo := collections.NewMap(state, string(payer.Bytes()))
	payerInfo.MustDelAt(service.Bytes())
	payerInfo.MustDelAt(getRevokeKey(service))
	payerInfo.MustDelAt(getLastOrdKey(service))
}

func getPublicKey(state kv.KVStoreReader, addr ledgerstate.Address, a assert.Assert) []byte {
	pkRegistry := collections.NewMapReadOnly(state, StateVarPublicKeys)
	ret, err := pkRegistry.GetAt(addr.Bytes())
	a.RequireNoError(err)
	return ret
}

func getRevokeKey(service ledgerstate.Address) []byte {
	return []byte(string(service.Bytes()) + "-revoke")
}

func getRevokeDeadline(nowis int64) time.Time {
	return time.Unix(0, nowis).Add(WarrantRevokePeriod)
}

func getLastOrdKey(service ledgerstate.Address) []byte {
	return []byte(string(service.Bytes()) + "-last")
}

func decodePayments(state kv.KVStoreReader, a assert.Assert) []*Payment {
	payments := collections.NewArray16ReadOnly(state, ParamPayments)
	n := payments.MustLen()
	a.Require(n > 0, "no payments found")

	ret := make([]*Payment, n)
	for i := range ret {
		data, err := payments.GetAt(uint16(i))
		a.RequireNoError(err)
		ret[i], err = NewPaymentFromBytes(data)
		a.RequireNoError(err)
	}
	return ret
}

func processPayments(ctx coretypes.Sandbox, payments []*Payment, payerAddr, targetAddr ledgerstate.Address, payerPubKey []byte) (uint64, []*Payment) {
	a := assert.NewAssert(ctx.Log())
	remainingWarrant, _, lastOrd := getWarrantInfoIntern(ctx.State(), payerAddr, targetAddr, a)
	a.Require(remainingWarrant > 0, "warrant == 0, can't settle payments")

	notSettled := make([]*Payment, 0)
	settledSum := uint64(0)
	for i, p := range payments {
		if uint64(p.Ord) <= lastOrd {
			// wrong order
			notSettled = append(notSettled, p)
			continue
		}
		data := paymentEssence(p.Ord, p.Amount, payerAddr, targetAddr)
		lastOrd = uint64(p.Ord)
		if !ctx.Utils().ED25519().ValidSignature(data, payerPubKey, p.SignatureShort) {
			ctx.Log().Infof("wrong signature")
			notSettled = append(notSettled, p)
			continue
		}
		if remainingWarrant < p.Amount {
			notSettled = append(notSettled, payments[i:]...)
			break
		}
		remainingWarrant -= p.Amount
		settledSum += p.Amount
		lastOrd = uint64(p.Ord)
	}
	if settledSum > 0 {
		tokens := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: settledSum})
		ctx.Send(targetAddr, tokens, nil)
	}
	payerInfo := collections.NewMap(ctx.State(), string(payerAddr.Bytes()))
	setWarrant(payerInfo, targetAddr, remainingWarrant)
	setLastOrd(payerInfo, targetAddr, lastOrd)
	return settledSum, notSettled
}
