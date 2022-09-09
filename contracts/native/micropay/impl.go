package micropay

// import (
// 	"time"

// 	iotago "github.com/iotaledger/iota.go/v3"
// 	"github.com/iotaledger/wasp/packages/isc"
// 	"github.com/iotaledger/wasp/packages/isc/assert"
// 	"github.com/iotaledger/wasp/packages/kv"
// 	"github.com/iotaledger/wasp/packages/kv/codec"
// 	"github.com/iotaledger/wasp/packages/kv/collections"
// 	"github.com/iotaledger/wasp/packages/kv/dict"
// 	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
// )

// var Processor = Contract.Processor(initialize,
// 	FuncPublicKey.WithHandler(publicKey),
// 	FuncAddWarrant.WithHandler(addWarrant),
// 	FuncRevokeWarrant.WithHandler(revokeWarrant),
// 	FuncCloseWarrant.WithHandler(closeWarrant),
// 	FuncSettle.WithHandler(settle),
// 	FuncGetChannelInfo.WithHandler(getWarrantInfo),
// )

// func initialize(_ isc.Sandbox) (dict.Dict, error) {
// 	return nil, nil
// }

// func publicKey(ctx isc.Sandbox) (dict.Dict, error) {
// 	panic("TODO implement")
// 	// a := assert.NewAssert(ctx.Log())
// 	// a.Requiref(ctx.Caller().Address().Type() != ledgerstate.AliasAddressType, "micropay.publicKey: caller must be an address")

// 	// par := kvdecoder.New(ctx.Params(), ctx.Log())

// 	// pubKeyBin := par.MustGetBytes(ParamPublicKey)
// 	// addr, err := ctx.Utils().ED25519().AddressFromPublicKey(pubKeyBin)
// 	// a.RequireNoError(err)
// 	// a.Requiref(addr.Equals(ctx.Caller().Address()), "public key does not correspond to the caller's address")

// 	// pkRegistry := collections.NewMap(ctx.State(), StateVarPublicKeys)
// 	// a.RequireNoError(pkRegistry.SetAt(addr.Bytes(), pubKeyBin))
// 	// return nil, nil
// }

// // addWarrant adds payment warrant for specific service address
// // Params:
// // - ParamServiceAddress iotago.Address
// func addWarrant(ctx isc.Sandbox) (dict.Dict, error) {
// 	panic("TODO implement")

// 	// par := kvdecoder.New(ctx.Params(), ctx.Log())
// 	// a := assert.NewAssert(ctx.Log())

// 	// a.Requiref(ctx.Caller().Address().Type() != ledgerstate.AliasAddressType, "micropay.addWarrant: caller must be an address")
// 	// payerAddr := ctx.Caller().Address()

// 	// a.Requiref(getPublicKey(ctx.State(), payerAddr, a) != nil,
// 	// 	fmt.Sprintf("unknown public key for address %s", payerAddr))

// 	// serviceAddr := par.MustGetAddress(ParamServiceAddress)
// 	// addWarrant := ctx.Allowance().BaseTokens
// 	// a.Requiref(addWarrant >= MinimumWarrantBaseTokens, fmt.Sprintf("warrant must be larger than %d base tokens", MinimumWarrantBaseTokens))

// 	// warrant, revoke, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, a)
// 	// a.Requiref(revoke == 0, fmt.Sprintf("warrant of %s for %s is being revoked", payerAddr, serviceAddr))

// 	// payerInfo := collections.NewMap(ctx.State(), string(payerAddr.Bytes()))
// 	// setWarrant(payerInfo, serviceAddr, warrant+addWarrant)

// 	// // all non-iota token accrue on-chain to the caller
// 	// // TODO refactor
// 	// sendBack := ctx.Allowance().Clone()
// 	// sendBack.Set(colored.IOTA, 0)

// 	// if len(sendBack) > 0 {
// 	// 	_, err := ctx.Call(
// 	// 		accounts.Contract.Hname(),
// 	// 		accounts.FuncDeposit.Hname(),
// 	// 		codec.MakeDict(map[string]interface{}{
// 	// 			accounts.ParamAgentID: ctx.Caller(),
// 	// 		}),
// 	// 		sendBack,
// 	// 	)

// 	// 	a.RequireNoError(err)
// 	// }

// 	// ctx.Event(fmt.Sprintf("[micropay.addWarrant] %s increased warrant %d -> %d i for %s",
// 	// 	payerAddr, warrant, warrant+addWarrant, serviceAddr))
// 	// return nil, nil
// }

// // revokeWarrant revokes payment warrant for specific service address
// // It will be in effect next 1 hour, the will be deleted
// // Params:
// // - ParamServiceAddress iotago.Address
// func revokeWarrant(ctx isc.Sandbox) (dict.Dict, error) {
// 	panic("TODO implement")

// 	// par := kvdecoder.New(ctx.Params(), ctx.Log())
// 	// a := assert.NewAssert(ctx.Log())

// 	// a.Requiref(ctx.Caller().Address().Type() != ledgerstate.AliasAddressType, "micropay.addWarrant: caller must be an address")
// 	// payerAddr := ctx.Caller().Address()
// 	// serviceAddr := par.MustGetAddress(ParamServiceAddress)

// 	// w, r, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, a)
// 	// a.Requiref(w > 0, fmt.Sprintf("warrant of %s to %s does not exist", payerAddr, serviceAddr))
// 	// a.Requiref(r == 0, fmt.Sprintf("warrant of %s to %s is already being revoked", payerAddr, serviceAddr))

// 	// revokeDeadline := getRevokeDeadline(ctx.Timestamp())
// 	// payerInfo := collections.NewMap(ctx.State(), string(payerAddr.Bytes()))
// 	// setWarrantRevoke(payerInfo, serviceAddr, revokeDeadline.Unix())

// 	// // send deterred request to self to revoke the warrant
// 	// iota1 := isc.NewFungibleBaseTokens(1)
// 	// meta := &isc.SendMetadata{
// 	// 	TargetContract: ctx.Contract(),
// 	// 	EntryPoint:     FuncCloseWarrant.Hname(),
// 	// 	Params: codec.MakeDict(map[string]interface{}{
// 	// 		ParamPayerAddress:   payerAddr,
// 	// 		ParamServiceAddress: serviceAddr,
// 	// 	}),
// 	// }
// 	// opts := isc.SendOptions{TimeLock: uint32(revokeDeadline.Unix())}
// 	// succ := ctx.Send(ctx.ChainID().AsAddress(), iota1, meta, opts)
// 	// a.Requiref(succ, "failed to issue deterred 'close warrant' request")
// 	// return nil, nil
// }

// // closeWarrant can only be sent from self. It closes the warrant account
// // - ParamServiceAddress iotago.Address
// // - ParamPayerAddress iotago.Address
// func closeWarrant(ctx isc.Sandbox) (dict.Dict, error) {
// 	panic("TODO implement")

// 	// a := assert.NewAssert(ctx.Log())
// 	// myAgentID := isc.NewAgentID(ctx.ChainID().AsAddress(), ctx.Contract())
// 	// a.Requiref(ctx.Caller().Equals(myAgentID), "caller must be self")

// 	// par := kvdecoder.New(ctx.Params(), ctx.Log())
// 	// payerAddr := par.MustGetAddress(ParamPayerAddress)
// 	// serviceAddr := par.MustGetAddress(ParamServiceAddress)
// 	// warrant, _, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, assert.NewAssert(ctx.Log()))
// 	// if warrant > 0 {
// 	// 	tokens := isc.NewFungibleBaseTokens(warrant)
// 	// 	succ := ctx.Send(payerAddr, tokens, nil)
// 	// 	a.Requiref(succ, "failed to send %d base tokens to address %s", warrant, payerAddr)
// 	// }
// 	// deleteWarrant(ctx.State(), payerAddr, serviceAddr)
// 	// return nil, nil
// }

// // Params:
// // - ParamPayerAddress address.address
// // - ParamPayments - array of encoded payments
// func settle(ctx isc.Sandbox) (dict.Dict, error) {
// 	panic("TODO implement")

// 	// a := assert.NewAssert(ctx.Log())
// 	// targetAddr := ctx.Caller().Address()
// 	// a.Requiref(targetAddr.Type() != ledgerstate.AliasAddressType, "micropay.addWarrant: caller must be an address")

// 	// par := kvdecoder.New(ctx.Params(), ctx.Log())
// 	// payerAddr := par.MustGetAddress(ParamPayerAddress)
// 	// payerPubKeyBin := getPublicKey(ctx.State(), payerAddr, a)
// 	// a.Requiref(payerPubKeyBin != nil, "public key unknown for %s", payerAddr)

// 	// payments := decodePayments(ctx.Params(), a)
// 	// settledSum, notSettled := processPayments(ctx, payments, payerAddr, targetAddr, payerPubKeyBin)
// 	// ctx.Event(fmt.Sprintf("[micropay.settle] settled %d i, num payments: %d, not settled payments: %d, payer: %s, target %s",
// 	// 	settledSum, len(payments)-len(notSettled), len(notSettled), payerAddr, targetAddr))
// 	// if len(notSettled) > 0 {
// 	// 	return nil, fmt.Errorf("number of payments failed to settle: %d", len(notSettled))
// 	// }
// 	// return nil, nil
// }

// // getWarrantInfo return warrant info for given payer and services addresses
// // Params:
// // - ParamServiceAddress iotago.Address
// // - ParamPayerAddress iotago.Address
// // Output:
// // - ParamWarrant int64 if == 0 no warrant
// // - ParamRevoked int64 is exists, timestamp in Unix nanosec when warrant will be revoked
// func getWarrantInfo(ctx isc.SandboxView) (dict.Dict, error) {
// 	// par := kvdecoder.New(ctx.Params(), ctx.Log())
// 	// payerAddr := par.MustGetAddress(ParamPayerAddress)
// 	// serviceAddr := par.MustGetAddress(ParamServiceAddress)
// 	// warrant, revoke, lastOrd := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, assert.NewAssert(ctx.Log()))
// 	// ret := dict.New()
// 	// if warrant > 0 {
// 	// 	ret.Set(ParamWarrant, codec.EncodeUint64(warrant))
// 	// }
// 	// if revoke > 0 {
// 	// 	ret.Set(ParamRevoked, codec.EncodeUint64(revoke))
// 	// }
// 	// if lastOrd > 0 {
// 	// 	ret.Set(ParamLastOrd, codec.EncodeUint64(lastOrd))
// 	// }
// 	// return ret, nil
// }

// func getWarrantInfoIntern(state kv.KVStoreReader, payer, service iotago.Address, a assert.Assert) (uint64, uint64, uint64) {
// 	payerInfo := collections.NewMapReadOnly(state, string(isc.BytesFromAddress(payer)))
// 	warrantBin := payerInfo.MustGetAt(isc.BytesFromAddress(service))
// 	warrant, err := codec.DecodeUint64(warrantBin, 0)
// 	a.RequireNoError(err)
// 	revokeBin := payerInfo.MustGetAt(getRevokeKey(service))
// 	revoke, err := codec.DecodeUint64(revokeBin, 0)
// 	a.RequireNoError(err)
// 	lastOrdBin := payerInfo.MustGetAt(getLastOrdKey(service))
// 	lastOrd, err := codec.DecodeUint64(lastOrdBin, 0)
// 	a.RequireNoError(err)
// 	return warrant, revoke, lastOrd
// }

// func setWarrant(payerAccount *collections.Map, service iotago.Address, value uint64) {
// 	payerAccount.MustSetAt(isc.BytesFromAddress(service), codec.EncodeUint64(value))
// }

// func setWarrantRevoke(payerAccount *collections.Map, service iotago.Address, deadline int64) {
// 	payerAccount.MustSetAt(getRevokeKey(service), codec.EncodeInt64(deadline))
// }

// func setLastOrd(payerAccount *collections.Map, service iotago.Address, lastOrd uint64) {
// 	payerAccount.MustSetAt(getLastOrdKey(service), codec.EncodeUint64(lastOrd))
// }

// func deleteWarrant(state kv.KVStore, payer, service iotago.Address) {
// 	payerInfo := collections.NewMap(state, string(isc.BytesFromAddress(payer)))
// 	payerInfo.MustDelAt(isc.BytesFromAddress(service))
// 	payerInfo.MustDelAt(getRevokeKey(service))
// 	payerInfo.MustDelAt(getLastOrdKey(service))
// }

// func getPublicKey(state kv.KVStoreReader, addr iotago.Address, a assert.Assert) []byte {
// 	pkRegistry := collections.NewMapReadOnly(state, StateVarPublicKeys)
// 	ret, err := pkRegistry.GetAt(isc.BytesFromAddress(addr))
// 	a.RequireNoError(err)
// 	return ret
// }

// func getRevokeKey(service iotago.Address) []byte {
// 	return []byte(string(isc.BytesFromAddress(service)) + "-revoke")
// }

// func getRevokeDeadline(nowis int64) time.Time {
// 	return time.Unix(0, nowis).Add(WarrantRevokePeriod)
// }

// func getLastOrdKey(service iotago.Address) []byte {
// 	return []byte(string(isc.BytesFromAddress(service)) + "-last")
// }

// func decodePayments(state kv.KVStoreReader, a assert.Assert) []*Payment {
// 	payments := collections.NewArray16ReadOnly(state, ParamPayments)
// 	n := payments.MustLen()
// 	a.Requiref(n > 0, "no payments found")

// 	ret := make([]*Payment, n)
// 	for i := range ret {
// 		data, err := payments.GetAt(uint16(i))
// 		a.RequireNoError(err)
// 		ret[i], err = NewPaymentFromBytes(data)
// 		a.RequireNoError(err)
// 	}
// 	return ret
// }

// func processPayments(ctx isc.Sandbox, payments []*Payment, payerAddr, targetAddr iotago.Address, payerPubKey []byte) (uint64, []*Payment) {
// 	a := assert.NewAssert(ctx.Log())
// 	remainingWarrant, _, lastOrd := getWarrantInfoIntern(ctx.State(), payerAddr, targetAddr, a)
// 	a.Requiref(remainingWarrant > 0, "warrant == 0, can't settle payments")

// 	notSettled := make([]*Payment, 0)
// 	settledSum := uint64(0)
// 	for i, p := range payments {
// 		if uint64(p.Ord) <= lastOrd {
// 			// wrong order
// 			notSettled = append(notSettled, p)
// 			continue
// 		}
// 		data := paymentEssence(p.Ord, p.Amount, payerAddr, targetAddr)
// 		lastOrd = uint64(p.Ord)
// 		if !ctx.Utils().ED25519().ValidSignature(data, payerPubKey, p.SignatureShort) {
// 			ctx.Log().Infof("wrong signature")
// 			notSettled = append(notSettled, p)
// 			continue
// 		}
// 		if remainingWarrant < p.Amount {
// 			notSettled = append(notSettled, payments[i:]...)
// 			break
// 		}
// 		remainingWarrant -= p.Amount
// 		settledSum += p.Amount
// 		lastOrd = uint64(p.Ord)
// 	}
// 	if settledSum > 0 {
// 		tokens := isc.NewFungibleTokens(settledSum, nil)
// 		ctx.Send(targetAddr, tokens, nil)
// 	}
// 	payerInfo := collections.NewMap(ctx.State(), string(isc.BytesFromAddress(payerAddr)))
// 	setWarrant(payerInfo, targetAddr, remainingWarrant)
// 	setLastOrd(payerInfo, targetAddr, lastOrd)
// 	return settledSum, notSettled
// }
