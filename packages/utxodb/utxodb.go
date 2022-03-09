package utxodb

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"golang.org/x/xerrors"
)

const (
	DefaultIOTASupply = iotago.TokenSupply

	Mi = 1_000_000

	// FundsFromFaucetAmount is how many iotas are returned from the faucet.
	FundsFromFaucetAmount = 1000 * Mi
)

var (
	genesisKeyPair = cryptolib.NewKeyPairFromSeed(cryptolib.NewSeedFromBytes([]byte("3.141592653589793238462643383279")))
	genesisAddress = genesisKeyPair.GetPublicKey().AsEd25519Address()
	genesisSigner  = iotago.NewInMemoryAddressSigner(genesisKeyPair.GetPrivateKey().AddressKeysForEd25519Address(genesisAddress))
)

type UnixSeconds uint64

// UtxoDB mocks the Tangle ledger by implementing a fully synchronous in-memory database
// of transactions. It ensures the consistency of the ledger and all added transactions
// by checking inputs, outputs and signatures.
type UtxoDB struct {
	mutex        sync.RWMutex
	supply       uint64
	seed         [cryptolib.SeedSize]byte
	l1Params     *parameters.L1
	transactions map[iotago.TransactionID]*iotago.Transaction
	utxo         map[iotago.OutputID]struct{}
	// latest milestone index and time. With each added transaction, global time moves
	// at least 1 ns and 1 milestone index. AdvanceClockBy advances the clock by duration and N milestone indices
	// globalLogicalTime can be ahead of real time due to AdvanceClockBy
	globalLogicalTime iscp.TimeData
	timeStep          time.Duration
}

type InitParams struct {
	initialTime time.Time
	timestep    time.Duration
	supply      uint64
	l1Params    *parameters.L1
	seed        [cryptolib.SeedSize]byte
}

func DefaultInitParams(seed ...[]byte) *InitParams {
	var seedBytes [cryptolib.SeedSize]byte
	if len(seed) > 0 {
		copy(seedBytes[:], seed[0])
	}
	return &InitParams{
		initialTime: time.Unix(1, 0),
		timestep:    1 * time.Millisecond,
		supply:      DefaultIOTASupply,
		l1Params:    parameters.L1ForTesting(),
		seed:        seedBytes,
	}
}

func (i *InitParams) WithInitialTime(t time.Time) *InitParams {
	i.initialTime = t
	return i
}

func (i *InitParams) WithTimeStep(timestep time.Duration) *InitParams {
	i.timestep = timestep
	return i
}

func (i *InitParams) WithL1Params(l1Params *parameters.L1) *InitParams {
	i.l1Params = l1Params
	return i
}

func (i *InitParams) WithSupply(supply uint64) *InitParams {
	i.supply = supply
	return i
}

// New creates a new UtxoDB instance
func New(params ...*InitParams) *UtxoDB {
	var p *InitParams
	if len(params) > 0 {
		p = params[0]
	} else {
		p = DefaultInitParams()
	}
	u := &UtxoDB{
		supply:       p.supply,
		seed:         p.seed,
		l1Params:     p.l1Params,
		transactions: make(map[iotago.TransactionID]*iotago.Transaction),
		utxo:         make(map[iotago.OutputID]struct{}),
		globalLogicalTime: iscp.TimeData{
			MilestoneIndex: 0,
			Time:           p.initialTime,
		},
		timeStep: p.timestep,
	}
	u.genesisInit()
	return u
}

func (u *UtxoDB) Seed() []byte {
	return u.seed[:]
}

func (u *UtxoDB) L1Params() *parameters.L1 {
	return u.l1Params
}

func (u *UtxoDB) RentStructure() *iotago.RentStructure {
	return u.l1Params.RentStructure()
}

func (u *UtxoDB) deSeriParams() *iotago.DeSerializationParameters {
	return u.l1Params.DeSerializationParameters
}

func (u *UtxoDB) genesisInit() {
	genesisTx, err := builder.NewTransactionBuilder(u.l1Params.NetworkID).
		AddInput(&builder.ToBeSignedUTXOInput{
			Address: genesisAddress,
			Output: &iotago.BasicOutput{
				Amount: DefaultIOTASupply,
				Conditions: iotago.UnlockConditions{
					&iotago.AddressUnlockCondition{Address: genesisAddress},
				},
			},
		}).
		AddOutput(&iotago.BasicOutput{
			Amount: DefaultIOTASupply,
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: genesisAddress},
			},
		}).
		Build(u.deSeriParams(), genesisSigner)
	if err != nil {
		panic(err)
	}
	u.addTransaction(genesisTx, true)
}

func (u *UtxoDB) addTransaction(tx *iotago.Transaction, isGenesis bool) {
	txid, err := tx.ID()
	if err != nil {
		panic(err)
	}
	// delete consumed outputs from the ledger
	inputs, err := u.getTransactionInputs(tx)
	if !isGenesis && err != nil {
		panic(err)
	}
	for outID := range inputs {
		delete(u.utxo, outID)
	}
	// store transaction
	u.transactions[*txid] = tx

	// add unspent outputs to the ledger
	for i := range tx.Essence.Outputs {
		utxo := &iotago.UTXOInput{
			TransactionID:          *txid,
			TransactionOutputIndex: uint16(i),
		}
		u.utxo[utxo.ID()] = struct{}{}
	}
	// advance clock
	u.advanceClockBy(u.timeStep, 1)
	u.checkLedgerBalance()
}

func (u *UtxoDB) advanceClockBy(step time.Duration, milestones uint32) {
	if milestones == 0 {
		panic("can't advance logical clock by 0 milestone indices")
	}
	if step == 0 {
		panic("can't advance clock by 0 nanoseconds")
	}
	u.globalLogicalTime.Time = u.globalLogicalTime.Time.Add(step)
	u.globalLogicalTime.MilestoneIndex += milestones
}

func (u *UtxoDB) AdvanceClockBy(step time.Duration, milestones uint32) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	u.advanceClockBy(step, milestones)
}

func (u *UtxoDB) GlobalTime() iscp.TimeData {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	return u.globalLogicalTime
}

func (u *UtxoDB) TimeStep() time.Duration {
	return u.timeStep
}

// GenesisAddress returns the genesis address.
func (u *UtxoDB) GenesisAddress() iotago.Address {
	return genesisAddress
}

func (u *UtxoDB) mustGetFundsFromFaucetTx(target iotago.Address, amount ...uint64) *iotago.Transaction {
	unspentOutputs := u.getUnspentOutputs(genesisAddress)
	if len(unspentOutputs) != 1 {
		panic("number of genesis outputs must be 1")
	}
	var inputOutput *iotago.BasicOutput
	var inputOutputID iotago.OutputID
	for oid, out := range unspentOutputs {
		inputOutput = out.(*iotago.BasicOutput)
		inputOutputID = oid
	}

	fundsAmount := uint64(FundsFromFaucetAmount)
	if len(amount) > 0 {
		fundsAmount = amount[0]
	}

	tx, err := builder.NewTransactionBuilder(u.l1Params.NetworkID).
		AddInput(&builder.ToBeSignedUTXOInput{
			Address:  genesisAddress,
			Output:   inputOutput,
			OutputID: inputOutputID,
		}).
		AddOutput(&iotago.BasicOutput{
			Amount: fundsAmount,
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: target},
			},
		}).
		AddOutput(&iotago.BasicOutput{
			Amount: inputOutput.Amount - fundsAmount,
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: genesisAddress},
			},
		}).
		Build(u.deSeriParams(), genesisSigner)
	if err != nil {
		panic(err)
	}
	return tx
}

// NewKeyPairByIndex deterministic private key
func (u *UtxoDB) NewKeyPairByIndex(index uint64) (*cryptolib.KeyPair, *iotago.Ed25519Address) {
	var tmp8 [8]byte
	binary.LittleEndian.PutUint64(tmp8[:], index)
	h := hashing.HashData(u.seed[:], tmp8[:])
	keyPair := cryptolib.NewKeyPairFromSeed(cryptolib.NewSeedFromBytes(h[:]))
	addr := keyPair.GetPublicKey().AsEd25519Address()
	return keyPair, addr
}

// GetFundsFromFaucet sends FundsFromFaucetAmount IOTA tokens from the genesis address to the given address.
func (u *UtxoDB) GetFundsFromFaucet(target iotago.Address, amount ...uint64) (*iotago.Transaction, error) {
	tx := u.mustGetFundsFromFaucetTx(target, amount...)
	return tx, u.AddToLedger(tx)
}

// Supply returns supply of the instance.
func (u *UtxoDB) Supply() uint64 {
	return u.supply
}

// GetOutput finds an output by ID (either spent or unspent).
func (u *UtxoDB) GetOutput(outID iotago.OutputID) iotago.Output {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.getOutput(outID)
}

func (u *UtxoDB) getOutput(outID iotago.OutputID) iotago.Output {
	tx, ok := u.getTransaction(outID.TransactionID())
	if !ok {
		return nil
	}
	if int(outID.Index()) >= len(tx.Essence.Outputs) {
		return nil
	}
	return tx.Essence.Outputs[outID.Index()]
}

func (u *UtxoDB) getTransactionInputs(tx *iotago.Transaction) (iotago.OutputSet, error) {
	inputs := iotago.OutputSet{}
	for _, input := range tx.Essence.Inputs {
		switch input.Type() {
		case iotago.InputUTXO:
			utxo := input.(*iotago.UTXOInput)
			out := u.getOutput(utxo.ID())
			if out == nil {
				return nil, xerrors.New("output not found")
			}
			inputs[utxo.ID()] = out
		case iotago.InputTreasury:
			panic("TODO")
		default:
			panic("unsupported input type")
		}
	}
	return inputs, nil
}

func (u *UtxoDB) validateTransaction(tx *iotago.Transaction) error {
	// serialize for syntactic check
	if _, err := tx.Serialize(serializer.DeSeriModePerformValidation, u.deSeriParams()); err != nil {
		return err
	}

	inputs, err := u.getTransactionInputs(tx)
	if err != nil {
		return xerrors.Errorf("validateTransaction: %w", err)
	}
	for outID := range inputs {
		if _, ok := u.utxo[outID]; !ok {
			return xerrors.New("referenced output is not unspent")
		}
	}

	{
		semValCtx := &iotago.SemanticValidationContext{
			ExtParas: &iotago.ExternalUnlockParameters{
				ConfMsIndex: u.globalLogicalTime.MilestoneIndex,
				ConfUnix:    uint32(u.globalLogicalTime.Time.Unix()),
			},
		}
		if err := tx.SemanticallyValidate(semValCtx, inputs); err != nil {
			return err
		}
	}

	return nil
}

// AddToLedger adds a transaction to UtxoDB, ensuring consistency of the UtxoDB ledger.
func (u *UtxoDB) AddToLedger(tx *iotago.Transaction) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if err := u.validateTransaction(tx); err != nil {
		return err
	}

	u.addTransaction(tx, false)
	return nil
}

// GetTransaction retrieves value transaction by its hash (ID).
func (u *UtxoDB) GetTransaction(txID iotago.TransactionID) (*iotago.Transaction, bool) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	return u.getTransaction(txID)
}

// MustGetTransaction same as GetTransaction only panics if transaction is not in UtxoDB.
func (u *UtxoDB) MustGetTransaction(txID iotago.TransactionID) *iotago.Transaction {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.mustGetTransaction(txID)
}

// GetUnspentOutputs returns all unspent outputs locked by the address with its ids
func (u *UtxoDB) GetUnspentOutputs(addr iotago.Address) (iotago.OutputSet, iotago.OutputIDs) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	outs := u.getUnspentOutputs(addr)

	ids := make(iotago.OutputIDs, len(outs))
	i := 0
	for id := range outs {
		ids[i] = id
		i++
	}

	return outs, ids
}

// GetAddressBalanceIotas returns the total amount of iotas owned by the address
func (u *UtxoDB) GetAddressBalanceIotas(addr iotago.Address) uint64 {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	ret := uint64(0)
	for _, out := range u.getUnspentOutputs(addr) {
		ret += out.Deposit()
	}
	return ret
}

// GetAddressBalances returns the total amount of iotas and tokens owned by the address
func (u *UtxoDB) GetAddressBalances(addr iotago.Address) *iscp.FungibleTokens {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	iotas := uint64(0)
	tokens := iotago.NativeTokenSum{}
	for _, out := range u.getUnspentOutputs(addr) {
		iotas += out.Deposit()
		tset, err := out.NativeTokenSet().Set()
		if err != nil {
			panic(err)
		}
		for _, token := range tset {
			val := tokens[token.ID]
			if val == nil {
				val = new(big.Int)
			}
			tokens[token.ID] = new(big.Int).Add(val, token.Amount)
		}
	}
	return iscp.FungibleTokensFromNativeTokenSum(iotas, tokens)
}

// GetAliasOutputs collects all outputs of type iotago.AliasOutput for the address
func (u *UtxoDB) GetAliasOutputs(addr iotago.Address) map[iotago.OutputID]*iotago.AliasOutput {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	outs := u.getUnspentOutputs(addr)
	ret := make(map[iotago.OutputID]*iotago.AliasOutput)
	for oid, out := range outs {
		if o, ok := out.(*iotago.AliasOutput); ok {
			ret[oid] = o
		}
	}
	return ret
}

func (u *UtxoDB) GetAddressNFTs(addr iotago.Address) map[iotago.OutputID]*iotago.NFTOutput {
	outs := u.getUnspentOutputs(addr)
	ret := make(map[iotago.OutputID]*iotago.NFTOutput)
	for oid, out := range outs {
		if o, ok := out.(*iotago.NFTOutput); ok {
			ret[oid] = o
		}
	}
	return ret
}

func (u *UtxoDB) getTransaction(txID iotago.TransactionID) (*iotago.Transaction, bool) {
	tx, ok := u.transactions[txID]
	if !ok {
		return nil, false
	}
	return tx, true
}

func (u *UtxoDB) mustGetTransaction(txID iotago.TransactionID) *iotago.Transaction {
	tx, ok := u.getTransaction(txID)
	if !ok {
		panic(fmt.Errorf("utxodb.mustGetTransaction: tx id doesn't exist: %s", txID))
	}
	return tx
}

func getOutputAddress(out iotago.Output, id *iotago.UTXOInput) iotago.Address {
	switch out := out.(type) {
	case iotago.TransIndepIdentOutput:
		return out.Ident()
	case iotago.TransDepIdentOutput:
		aliasID := out.Chain().(iotago.AliasID)
		if aliasID.Empty() {
			aliasID = iotago.AliasIDFromOutputID(id.ID())
		}
		return aliasID.ToAddress()
	default:
		panic("unknown ident output type")
	}
}

func (u *UtxoDB) getUnspentOutputs(addr iotago.Address) iotago.OutputSet {
	ret := make(iotago.OutputSet)
	for oid := range u.utxo {
		out := u.getOutput(oid)
		if getOutputAddress(out, oid.UTXOInput()).Equal(addr) {
			ret[oid] = out
		}
	}
	return ret
}

func (u *UtxoDB) checkLedgerBalance() {
	total := uint64(0)
	for outID := range u.utxo {
		out := u.getOutput(outID)
		total += out.Deposit()
	}
	if total != u.Supply() {
		panic("utxodb: wrong ledger balance")
	}
}
