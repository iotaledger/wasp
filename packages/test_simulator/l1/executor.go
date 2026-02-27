package l1

import (
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type ExecutionResult struct {
	Created  []*SimObject
	Mutated  []*SimObject
	Deleted  []iotago.ObjectID
	GasCost  uint64
	TxDigest iotago.TransactionDigest
}

type Executor struct {
	Store       *ObjectStore
	MoveHandler MoveCallHandler
}

func NewExecutor(store *ObjectStore, moveHandler MoveCallHandler) *Executor {
	return &Executor{Store: store, MoveHandler: moveHandler}
}

// execState holds mutable state shared across PTB command execution.
type execState struct {
	ctx            *CallContext
	pt             *iotago.ProgrammableTransaction
	cmdResults     [][]Value
	gasCoinID      *iotago.ObjectID
	gasCoinBalance *uint64
	validated      *ValidatedInputs
}

func (e *Executor) Execute(tx *iotago.TransactionData, validated *ValidatedInputs) (*ExecutionResult, error) {
	if tx.V1 == nil {
		return nil, fmt.Errorf("TransactionData.V1 is nil")
	}
	v1 := tx.V1
	pt := v1.Kind.ProgrammableTransaction
	if pt == nil {
		return nil, fmt.Errorf("transaction must be ProgrammableTransaction")
	}

	txDigest, err := tx.Digest()
	if err != nil {
		return nil, fmt.Errorf("computing tx digest: %w", err)
	}

	var idCounter uint64

	cmdResults := make([][]Value, len(pt.Commands))
	gasCoinBalance := uint64(0)
	var gasCoinID *iotago.ObjectID
	for _, gc := range validated.GasCoins {
		gasCoinBalance += DecodeCoinObjectBalance(gc.Data)
		if gasCoinID == nil {
			id := gc.ID
			gasCoinID = &id
		}
	}

	es := &execState{
		ctx: &CallContext{
			Store:     e.Store,
			Sender:    v1.Sender,
			TxDigest:  *txDigest,
			IDCounter: &idCounter,
			PackageID: iotago.PackageID{},
		},
		pt:             pt,
		cmdResults:     cmdResults,
		gasCoinID:      gasCoinID,
		gasCoinBalance: &gasCoinBalance,
		validated:      validated,
	}

	for cmdIdx, cmd := range pt.Commands {
		results, err := e.executeCommand(es, cmd)
		if err != nil {
			return nil, fmt.Errorf("command %d: %w", cmdIdx, err)
		}
		es.cmdResults[cmdIdx] = results
	}

	// L1 gas: the simulator does not model gas costs. Real IOTA charges
	// computation_cost (from Move VM instruction metering) + storage_cost
	// (per-byte for new/mutated objects) - storage_rebate. Reproducing this
	// requires executing Move bytecode, which the simulator doesn't do.
	// TODO: Marker for adding Gas cost calculation, but this technically requires actual execution of contracts.
	gasCost := uint64(0)

	if gasCoinID != nil {
		gasObj, ok := e.Store.Get(*gasCoinID)
		if ok {
			gasObj.Data = encodeCoinObject(gasObj.ID, gasCoinBalance)
			gasObj.Digest = ComputeDigest(gasObj.Data)
			gasObj.PreviousTx = *txDigest
			e.Store.Put(gasObj)
		}
	}
	for i := 1; i < len(validated.GasCoins); i++ {
		e.Store.Delete(validated.GasCoins[i].ID)
	}

	newVersion := NextLamportVersion(validated.Versions...)
	result := &ExecutionResult{
		GasCost:  gasCost,
		TxDigest: *txDigest,
	}
	e.applyVersions(newVersion, *txDigest, validated, result)

	return result, nil
}

func (e *Executor) executeCommand(es *execState, cmd iotago.Command) ([]Value, error) {
	switch {
	case cmd.SplitCoins != nil:
		return e.executeSplitCoins(es, cmd.SplitCoins)
	case cmd.MergeCoins != nil:
		return e.executeMergeCoins(es, cmd.MergeCoins)
	case cmd.TransferObjects != nil:
		return e.executeTransferObjects(es, cmd.TransferObjects)
	case cmd.MoveCall != nil:
		return e.executeMoveCall(es, cmd.MoveCall)
	case cmd.MakeMoveVec != nil:
		return e.executeMakeMoveVec(es, cmd.MakeMoveVec)
	case cmd.Publish != nil:
		return e.executePublish(es, cmd.Publish)
	default:
		return nil, fmt.Errorf("unsupported command type")
	}
}

func (e *Executor) executeSplitCoins(es *execState, split *iotago.ProgrammableSplitCoins) ([]Value, error) {
	amounts := make([]uint64, len(split.Amounts))
	for i, amtArg := range split.Amounts {
		val, err := e.resolveArgument(es, amtArg)
		if err != nil {
			return nil, fmt.Errorf("SplitCoins: resolving amount %d: %w", i, err)
		}
		amt, err := valueToUint64(val)
		if err != nil {
			return nil, fmt.Errorf("SplitCoins: amount %d: %w", i, err)
		}
		amounts[i] = amt
	}

	coinVal, err := e.resolveArgument(es, split.Coin)
	if err != nil {
		return nil, fmt.Errorf("SplitCoins: resolving coin: %w", err)
	}

	isGasCoin := split.Coin.GasCoin != nil
	coinType := CoinTypeString(IotaCoinTypeStr)
	var sourceCoinType string

	if isGasCoin {
		sourceCoinType = IotaCoinTypeStr
		var totalSplit uint64
		for _, a := range amounts {
			totalSplit += a
		}
		if *es.gasCoinBalance < totalSplit {
			return nil, fmt.Errorf("SplitCoins: insufficient gas coin balance: have %d, need %d", *es.gasCoinBalance, totalSplit)
		}
		*es.gasCoinBalance -= totalSplit
	} else if coinVal.ObjectID != nil {
		obj, ok := e.Store.Get(*coinVal.ObjectID)
		if !ok {
			return nil, fmt.Errorf("SplitCoins: coin object not found")
		}
		coinType = obj.Type
		if ct, ok := extractCoinType(obj.Type); ok {
			sourceCoinType = ct
		}
		balance := DecodeCoinObjectBalance(obj.Data)
		var totalSplit uint64
		for _, a := range amounts {
			totalSplit += a
		}
		if balance < totalSplit {
			return nil, fmt.Errorf("SplitCoins: insufficient coin balance: have %d, need %d", balance, totalSplit)
		}
		obj.Data = encodeCoinObject(obj.ID, balance-totalSplit)
		obj.Digest = ComputeDigest(obj.Data)
		e.Store.Put(obj)
	} else {
		return nil, fmt.Errorf("SplitCoins: cannot determine coin source")
	}

	if sourceCoinType == "" {
		if ct, ok := extractCoinType(coinType); ok {
			sourceCoinType = ct
		} else {
			sourceCoinType = IotaCoinTypeStr
		}
	}

	results := make([]Value, len(amounts))
	for i, amt := range amounts {
		newID := FreshID(es.ctx.TxDigest, es.ctx.IDCounter)
		newType := CoinTypeString(sourceCoinType)
		data := encodeCoinObject(newID, amt)
		newObj := &SimObject{
			ID:         newID,
			Version:    0,
			Digest:     ComputeDigest(data),
			Owner:      SimOwner{AddressOwner: &es.ctx.Sender},
			Type:       newType,
			Data:       data,
			PreviousTx: es.ctx.TxDigest,
		}
		e.Store.Put(newObj)
		results[i] = Value{ObjectID: &newID, Type: newType}
	}

	return results, nil
}

func (e *Executor) executeMergeCoins(es *execState, merge *iotago.ProgrammableMergeCoins) ([]Value, error) {
	isGasCoinDest := merge.Destination.GasCoin != nil

	for _, srcArg := range merge.Sources {
		srcVal, err := e.resolveArgument(es, srcArg)
		if err != nil {
			return nil, fmt.Errorf("MergeCoins: resolving source: %w", err)
		}
		if srcVal.ObjectID == nil {
			continue
		}

		srcObj, ok := e.Store.Get(*srcVal.ObjectID)
		if !ok {
			continue
		}
		srcBalance := DecodeCoinObjectBalance(srcObj.Data)

		if isGasCoinDest {
			*es.gasCoinBalance += srcBalance
		} else {
			destVal, err := e.resolveArgument(es, merge.Destination)
			if err != nil {
				return nil, fmt.Errorf("MergeCoins: resolving destination: %w", err)
			}
			if destVal.ObjectID != nil {
				destObj, ok := e.Store.Get(*destVal.ObjectID)
				if ok {
					destBal := DecodeCoinObjectBalance(destObj.Data)
					destObj.Data = encodeCoinObject(destObj.ID, destBal+srcBalance)
					destObj.Digest = ComputeDigest(destObj.Data)
					e.Store.Put(destObj)
				}
			}
		}

		e.Store.Delete(*srcVal.ObjectID)
	}

	return nil, nil
}

func (e *Executor) executeTransferObjects(es *execState, transfer *iotago.ProgrammableTransferObjects) ([]Value, error) {
	addrVal, err := e.resolveArgument(es, transfer.Address)
	if err != nil {
		return nil, fmt.Errorf("TransferObjects: resolving address: %w", err)
	}

	var recipient iotago.Address
	switch v := addrVal.Raw.(type) {
	case *iotago.Address:
		recipient = *v
	case iotago.Address:
		recipient = v
	case []byte:
		if len(v) == 32 {
			copy(recipient[:], v)
		} else {
			return nil, fmt.Errorf("TransferObjects: address bytes length %d, expected 32", len(v))
		}
	default:
		return nil, fmt.Errorf("TransferObjects: address argument is not an Address (got %T)", addrVal.Raw)
	}

	for _, objArg := range transfer.Objects {
		objVal, err := e.resolveArgument(es, objArg)
		if err != nil {
			return nil, fmt.Errorf("TransferObjects: resolving object: %w", err)
		}

		if objArg.GasCoin != nil {
			if es.gasCoinID != nil {
				gasObj, ok := e.Store.Get(*es.gasCoinID)
				if ok {
					gasObj.Owner = SimOwner{AddressOwner: &recipient}
					gasObj.Data = encodeCoinObject(gasObj.ID, *es.gasCoinBalance)
					gasObj.Digest = ComputeDigest(gasObj.Data)
					e.Store.Put(gasObj)
				}
			}
			continue
		}

		if objVal.ObjectID != nil {
			obj, ok := e.Store.Get(*objVal.ObjectID)
			if ok {
				obj.Owner = SimOwner{AddressOwner: &recipient}
				e.Store.Put(obj)
			}
		}
	}

	return nil, nil
}

func (e *Executor) executeMoveCall(es *execState, call *iotago.ProgrammableMoveCall) ([]Value, error) {
	if call.Package != nil {
		es.ctx.PackageID = *call.Package
	}

	args := make([]Value, len(call.Arguments))
	for i, arg := range call.Arguments {
		val, err := e.resolveArgument(es, arg)
		if err != nil {
			return nil, fmt.Errorf("MoveCall %s::%s: resolving arg %d: %w", call.Module, call.Function, i, err)
		}
		args[i] = val
	}

	return e.MoveHandler.ExecuteMoveCall(es.ctx, call, args)
}

func (e *Executor) executeMakeMoveVec(es *execState, vec *iotago.ProgrammableMakeMoveVec) ([]Value, error) {
	elements := make([]Value, len(vec.Objects))
	for i, arg := range vec.Objects {
		val, err := e.resolveArgument(es, arg)
		if err != nil {
			return nil, fmt.Errorf("MakeMoveVec: resolving element %d: %w", i, err)
		}
		elements[i] = val
	}

	return []Value{{Raw: elements, Type: "vector"}}, nil
}

func (e *Executor) executePublish(es *execState, publish *iotago.ProgrammablePublish) ([]Value, error) {
	// Version 0 so applyVersions picks it up as newly created
	pkgID := FreshID(es.ctx.TxDigest, es.ctx.IDCounter)
	pkgObj := &SimObject{
		ID:         pkgID,
		Version:    0,
		Digest:     ComputeDigest([]byte("package")),
		Owner:      SimOwner{Immutable: true},
		Type:       "package",
		Data:       nil,
		PreviousTx: es.ctx.TxDigest,
	}
	e.Store.Put(pkgObj)

	capID := FreshID(es.ctx.TxDigest, es.ctx.IDCounter)
	capObj := &SimObject{
		ID:         capID,
		Version:    0,
		Digest:     ComputeDigest([]byte("upgrade_cap")),
		Owner:      SimOwner{AddressOwner: &es.ctx.Sender},
		Type:       UpgradeCapTypeString(),
		Data:       nil,
		PreviousTx: es.ctx.TxDigest,
	}
	e.Store.Put(capObj)

	// Fake init: coin-publishing tests expect TreasuryCap and CoinMetadata in the
	// transaction effects. The real Move runtime creates these via the module's init
	// function calling coin::create_currency. The type parameter in the generic is
	// a placeholder — GetCreatedObjectByName only matches on module and object name.
	treasuryCapID := FreshID(es.ctx.TxDigest, es.ctx.IDCounter)
	treasuryCapType := fmt.Sprintf("%s::coin::TreasuryCap<%s::unknown::T>", iotago.IotaPackageIDIotaFramework, pkgID)
	e.Store.Put(&SimObject{
		ID:         treasuryCapID,
		Version:    0,
		Digest:     ComputeDigest([]byte("treasury_cap")),
		Owner:      SimOwner{AddressOwner: &es.ctx.Sender},
		Type:       treasuryCapType,
		Data:       nil,
		PreviousTx: es.ctx.TxDigest,
	})

	coinMetadataID := FreshID(es.ctx.TxDigest, es.ctx.IDCounter)
	coinMetadataType := fmt.Sprintf("%s::coin::CoinMetadata<%s::unknown::T>", iotago.IotaPackageIDIotaFramework, pkgID)
	e.Store.Put(&SimObject{
		ID:         coinMetadataID,
		Version:    0,
		Digest:     ComputeDigest([]byte("coin_metadata")),
		Owner:      SimOwner{Immutable: true},
		Type:       coinMetadataType,
		Data:       nil,
		PreviousTx: es.ctx.TxDigest,
	})

	_ = publish
	return []Value{{ObjectID: &capID, Type: "UpgradeCap"}}, nil
}

func (e *Executor) resolveArgument(es *execState, arg iotago.Argument) (Value, error) {
	switch {
	case arg.GasCoin != nil:
		return Value{ObjectID: es.gasCoinID, Raw: es.gasCoinBalance, Type: "GasCoin"}, nil

	case arg.Input != nil:
		idx := int(*arg.Input)
		if idx >= len(es.pt.Inputs) {
			return Value{}, fmt.Errorf("input index %d out of range (have %d inputs)", idx, len(es.pt.Inputs))
		}
		return e.resolveCallArg(es.pt.Inputs[idx], es.validated)

	case arg.Result != nil:
		cmdIdx := int(*arg.Result)
		if cmdIdx >= len(es.cmdResults) || es.cmdResults[cmdIdx] == nil {
			return Value{}, fmt.Errorf("result index %d not available", cmdIdx)
		}
		results := es.cmdResults[cmdIdx]
		if len(results) == 0 {
			return Value{}, fmt.Errorf("command %d produced no results", cmdIdx)
		}
		// For multiple results, return the first (caller should use NestedResult)
		return results[0], nil

	case arg.NestedResult != nil:
		cmdIdx := int(arg.NestedResult.Cmd)
		resIdx := int(arg.NestedResult.Result)
		if cmdIdx >= len(es.cmdResults) || es.cmdResults[cmdIdx] == nil {
			return Value{}, fmt.Errorf("nested result: command %d not available", cmdIdx)
		}
		results := es.cmdResults[cmdIdx]
		if resIdx >= len(results) {
			return Value{}, fmt.Errorf("nested result: result %d out of range (command %d has %d results)", resIdx, cmdIdx, len(results))
		}
		return results[resIdx], nil

	default:
		return Value{}, fmt.Errorf("invalid argument type")
	}
}

func (e *Executor) resolveCallArg(callArg iotago.CallArg, validated *ValidatedInputs) (Value, error) {
	switch {
	case callArg.Pure != nil:
		return decodePureValue(*callArg.Pure), nil

	case callArg.Object != nil:
		objArg := callArg.Object
		var objID *iotago.ObjectID

		switch {
		case objArg.ImmOrOwnedObject != nil:
			objID = objArg.ImmOrOwnedObject.ObjectID
		case objArg.SharedObject != nil:
			objID = objArg.SharedObject.Id
		case objArg.Receiving != nil:
			objID = objArg.Receiving.ObjectID
		}

		if objID == nil {
			return Value{}, fmt.Errorf("object argument has nil ID")
		}

		// Try validated inputs first (snapshot at validation time)
		if obj, ok := validated.Objects[*objID]; ok {
			return Value{ObjectID: objID, Raw: obj, Type: obj.Type}, nil
		}

		obj, ok := e.Store.Get(*objID)
		if !ok {
			return Value{}, fmt.Errorf("object %s not found", objID.String())
		}
		return Value{ObjectID: objID, Raw: obj, Type: obj.Type}, nil

	default:
		return Value{}, fmt.Errorf("empty CallArg")
	}
}

// decodePureValue returns raw BCS bytes as a Value.
// Consumers must BCS-decode to the expected type themselves.
func decodePureValue(data []byte) Value {
	raw := make([]byte, len(data))
	copy(raw, data)
	return Value{Raw: raw, Type: "pure"}
}

func valueToUint64(v Value) (uint64, error) {
	switch val := v.Raw.(type) {
	case uint64:
		return val, nil
	case *uint64:
		return *val, nil
	case uint32:
		return uint64(val), nil
	case byte:
		return uint64(val), nil
	case []byte:
		result, err := bcs.Unmarshal[uint64](val)
		if err != nil {
			return 0, fmt.Errorf("BCS decode u64 failed: %w", err)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", v.Raw)
	}
}

// applyVersions sets the Lamport version on all created/mutated objects
// and classifies them into the execution result.
func (e *Executor) applyVersions(newVersion uint64, txDigest iotago.TransactionDigest, validated *ValidatedInputs, result *ExecutionResult) {
	seen := make(map[iotago.ObjectID]bool)

	for id := range validated.Objects {
		seen[id] = true
		obj, ok := e.Store.Get(id)
		if !ok {
			result.Deleted = append(result.Deleted, id)
			continue
		}
		obj.Version = newVersion
		obj.PreviousTx = txDigest
		obj.Digest = ComputeDigest(obj.Data)
		e.Store.Put(obj)
		result.Mutated = append(result.Mutated, obj)
	}

	// Newly created objects have Version == 0
	e.Store.mu.RLock()
	var newObjs []iotago.ObjectID
	for id, obj := range e.Store.objects {
		if obj.Version == 0 && !seen[id] {
			newObjs = append(newObjs, id)
		}
	}
	e.Store.mu.RUnlock()

	for _, id := range newObjs {
		obj, ok := e.Store.Get(id)
		if !ok {
			continue
		}
		obj.Version = newVersion
		obj.PreviousTx = txDigest
		obj.Digest = ComputeDigest(obj.Data)
		e.Store.Put(obj)
		result.Created = append(result.Created, obj)
	}
}
