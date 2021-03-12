package micropay

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"time"
)

func initialize(_ coretypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

func publicKey(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(ctx.Caller().IsAddress(), "micropay.publicKey: caller must be an address")

	par := kvdecoder.New(ctx.Params(), ctx.Log())

	pubKeyBin := par.MustGetBytes(ParamPublicKey)
	addr, err := ctx.Utils().ED25519().AddressFromPublicKey(pubKeyBin)
	a.RequireNoError(err)
	a.Require(addr == ctx.Caller().MustAddress(), "public key does not correspond to the caller's address")

	pkRegistry := collections.NewMap(ctx.State(), StateVarPublicKeys)
	a.RequireNoError(pkRegistry.SetAt(addr[:], pubKeyBin))
	return nil, nil
}

// addWarrant adds payment warrant for specific service address
// Params:
// - ParamServiceAddress address.Address
func addWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Require(ctx.Caller().IsAddress(), "payer must be an address")
	payerAddr := ctx.Caller().MustAddress()

	a.Require(getPublicKey(ctx.State(), payerAddr, a) != nil,
		fmt.Sprintf("unknown public key for address %s", payerAddr))

	serviceAddr := par.MustGetAddress(ParamServiceAddress)
	addWarrant := ctx.IncomingTransfer().Balance(balance.ColorIOTA)
	a.Require(addWarrant >= MinimumWarrantIotas, fmt.Sprintf("warrant must be larger than %d iotas", MinimumWarrantIotas))

	warrant, revoke, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, a)
	a.Require(revoke == 0, fmt.Sprintf("warrant of %s for %s is being revoked", payerAddr, serviceAddr))

	payerInfo := collections.NewMap(ctx.State(), string(payerAddr[:]))
	setWarrant(payerInfo, serviceAddr, warrant+addWarrant)

	// all non-iota token accrue on-chain to the caller
	sendBack := ctx.IncomingTransfer().TakeOutColor(balance.ColorIOTA)
	err := accounts.Accrue(ctx, ctx.Caller(), sendBack)
	a.RequireNoError(err)

	ctx.Event(fmt.Sprintf("[micropay.addWarrant] %s increased warrant %d -> %d i for %s",
		payerAddr, warrant, warrant+addWarrant, serviceAddr))
	return nil, nil
}

// revokeWarrant revokes payment warrant for specific service address
// It will be in effect next 1 hour, the will be deleted
// Params:
// - ParamServiceAddress address.Address
func revokeWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	a.Require(ctx.Caller().IsAddress(), "payer must be an address")
	payerAddr := ctx.Caller().MustAddress()
	serviceAddr := par.MustGetAddress(ParamServiceAddress)

	w, r, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, a)
	a.Require(w > 0, fmt.Sprintf("warrant of %s to %s does not exist", payerAddr, serviceAddr))
	a.Require(r == 0, fmt.Sprintf("warrant of %s to %s is already being revoked", payerAddr, serviceAddr))

	revokeDeadline := getRevokeDeadline(ctx.GetTimestamp())
	payerInfo := collections.NewMap(ctx.State(), string(payerAddr[:]))
	setWarrantRevoke(payerInfo, serviceAddr, revokeDeadline.Unix())

	succ := ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: ctx.ContractID(),
		EntryPoint:       coretypes.Hn(FuncCloseWarrant),
		TimeLock:         uint32(revokeDeadline.Unix()),
		Params: codec.MakeDict(map[string]interface{}{
			ParamPayerAddress:   payerAddr,
			ParamServiceAddress: serviceAddr,
		}),
	})
	a.Require(succ, "failed to post time-locked 'closeWarrant' request to self")
	return nil, nil
}

// closeWarrant can only be sent from self. It closes the warrant account
// - ParamServiceAddress address.Address
// - ParamPayerAddress address.Address
func closeWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(ctx.Caller() == coretypes.NewAgentIDFromContractID(ctx.ContractID()), "caller must be self")

	par := kvdecoder.New(ctx.Params(), ctx.Log())
	payerAddr := par.MustGetAddress(ParamPayerAddress)
	serviceAddr := par.MustGetAddress(ParamServiceAddress)
	warrant, _, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, assert.NewAssert(ctx.Log()))
	if warrant > 0 {
		succ := ctx.TransferToAddress(payerAddr, cbalances.NewIotasOnly(warrant))
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
	a.Require(ctx.Caller().IsAddress(), "caller must be an address")
	targetAddr := ctx.Caller().MustAddress()

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
// - ParamServiceAddress address.Address
// - ParamPayerAddress address.Address
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
		ret.Set(ParamWarrant, codec.EncodeInt64(warrant))
	}
	if revoke > 0 {
		ret.Set(ParamRevoked, codec.EncodeInt64(revoke))
	}
	if lastOrd > 0 {
		ret.Set(ParamLastOrd, codec.EncodeInt64(lastOrd))
	}
	return ret, nil
}

//  utility

func getWarrantInfoIntern(state kv.KVStoreReader, payer, service address.Address, a assert.Assert) (int64, int64, int64) {
	payerInfo := collections.NewMapReadOnly(state, string(payer[:]))
	warrantBin, err := payerInfo.GetAt(service[:])
	a.RequireNoError(err)
	warrant, exists, err := codec.DecodeInt64(warrantBin)
	a.RequireNoError(err)
	if !exists {
		warrant = 0
	}
	revokeBin, err := payerInfo.GetAt(getRevokeKey(service))
	revoke, exists, err := codec.DecodeInt64(revokeBin)
	if !exists {
		revoke = 0
	}
	lastOrdBin, err := payerInfo.GetAt(getLastOrdKey(service))
	lastOrd, exists, err := codec.DecodeInt64(lastOrdBin)
	if !exists {
		lastOrd = 0
	}
	return warrant, revoke, lastOrd
}

func setWarrant(payerAccount *collections.Map, service address.Address, value int64) {
	payerAccount.MustSetAt(service[:], codec.EncodeInt64(value))
}

func setWarrantRevoke(payerAccount *collections.Map, service address.Address, deadline int64) {
	payerAccount.MustSetAt(getRevokeKey(service), codec.EncodeInt64(deadline))
}

func setLastOrd(payerAccount *collections.Map, service address.Address, lastOrd int64) {
	payerAccount.MustSetAt(getLastOrdKey(service), codec.EncodeInt64(lastOrd))
}

func deleteWarrant(state kv.KVStore, payer, service address.Address) {
	payerInfo := collections.NewMap(state, string(payer[:]))
	payerInfo.MustDelAt(service[:])
	payerInfo.MustDelAt(getRevokeKey(service))
	payerInfo.MustDelAt(getLastOrdKey(service))
}

func getPublicKey(state kv.KVStoreReader, addr address.Address, a assert.Assert) []byte {
	pkRegistry := collections.NewMapReadOnly(state, StateVarPublicKeys)
	ret, err := pkRegistry.GetAt(addr[:])
	a.RequireNoError(err)
	return ret
}

func getRevokeKey(service address.Address) []byte {
	return []byte(string(service[:]) + "-revoke")
}

func getRevokeDeadline(nowis int64) time.Time {
	return time.Unix(0, nowis).Add(WarrantRevokePeriod)
}

func getLastOrdKey(service address.Address) []byte {
	return []byte(string(service[:]) + "-last")
}

func decodePayments(state kv.KVStoreReader, a assert.Assert) []*Payment {
	payments := collections.NewArrayReadOnly(state, ParamPayments)
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

func processPayments(ctx coretypes.Sandbox, payments []*Payment, payerAddr, targetAddr address.Address, payerPubKey []byte) (int64, []*Payment) {
	a := assert.NewAssert(ctx.Log())
	remainingWarrant, _, lastOrd := getWarrantInfoIntern(ctx.State(), payerAddr, targetAddr, a)
	a.Require(remainingWarrant > 0, "warrant == 0, can't settle payments")

	notSettled := make([]*Payment, 0)
	settledSum := int64(0)
	for i, p := range payments {
		if int64(p.Ord) <= lastOrd {
			// wrong order
			notSettled = append(notSettled, p)
			continue
		}
		data := paymentEssence(p.Ord, p.Amount, payerAddr, targetAddr)
		lastOrd = int64(p.Ord)
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
		lastOrd = int64(p.Ord)
	}
	if settledSum > 0 {
		ctx.TransferToAddress(targetAddr, cbalances.NewIotasOnly(settledSum))
	}
	payerInfo := collections.NewMap(ctx.State(), string(payerAddr[:]))
	setWarrant(payerInfo, targetAddr, remainingWarrant)
	setLastOrd(payerInfo, targetAddr, lastOrd)
	return settledSum, notSettled
}
