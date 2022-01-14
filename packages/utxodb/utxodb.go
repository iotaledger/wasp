package utxodb

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

const (
	DefaultIOTASupply = iotago.TokenSupply

	Mi = 1_000_000

	// FundsFromFaucetAmount is how many iotas are returned from the faucet.
	FundsFromFaucetAmount = 1 * Mi
)

var (
	genesisKeyPair = cryptolib.NewKeyPairFromSeed(cryptolib.SeedFromByteArray([]byte("3.141592653589793238462643383279")))
	genesisAddress = cryptolib.Ed25519AddressFromPubKey(genesisKeyPair.PublicKey)
	genesisSigner  = iotago.NewInMemoryAddressSigner(iotago.NewAddressKeysForEd25519Address(genesisAddress, genesisKeyPair.PrivateKey))
)

type UnixSeconds uint64

// UtxoDB mocks the Tangle ledger by implementing a fully synchronous in-memory database
// of transactions. It ensures the consistency of the ledger and all added transactions
// by checking inputs, outputs and signatures.
type UtxoDB struct {
	mutex         sync.RWMutex
	supply        uint64
	seed          [cryptolib.SeedSize]byte
	rentStructure *iotago.RentStructure
	transactions  map[iotago.TransactionID]*iotago.Transaction
	utxo          map[iotago.OutputID]struct{}
	// latest milestone index and time. With each added transaction, global time moves
	// at least 1 ns and 1 milestone index. AdvanceClockBy advances the clock by duration and N milestone indices
	// globalLogicalTime can be ahead of real time due to AdvanceClockBy
	globalLogicalTime iscp.TimeData
	timeStep          time.Duration
}

type InitParams struct {
	timestep      time.Duration
	supply        uint64
	rentStructure *iotago.RentStructure
	seed          [cryptolib.SeedSize]byte
}

func DefaultInitParams(seed ...[]byte) *InitParams {
	var seedBytes [cryptolib.SeedSize]byte
	if len(seed) > 0 {
		copy(seedBytes[:], seed[0])
	}
	return &InitParams{
		timestep:      1 * time.Millisecond,
		supply:        DefaultIOTASupply,
		rentStructure: &iotago.RentStructure{},
		seed:          seedBytes,
	}
}

func (i *InitParams) WithTimeStep(timestep time.Duration) *InitParams {
	i.timestep = timestep
	return i
}

func (i *InitParams) WithRentStructure(r *iotago.RentStructure) *InitParams {
	i.rentStructure = r
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
		supply:        p.supply,
		seed:          p.seed,
		rentStructure: p.rentStructure,
		transactions:  make(map[iotago.TransactionID]*iotago.Transaction),
		utxo:          make(map[iotago.OutputID]struct{}),
		globalLogicalTime: iscp.TimeData{
			MilestoneIndex: 0,
			Time:           time.Unix(1, 0),
		},
		timeStep: p.timestep,
	}
	u.genesisInit()
	return u
}

func (u *UtxoDB) Seed() []byte {
	return u.seed[:]
}

func (u *UtxoDB) RentStructure() *iotago.RentStructure {
	return u.rentStructure
}

func (u *UtxoDB) deSeriParams() *iotago.DeSerializationParameters {
	return &iotago.DeSerializationParameters{RentStructure: u.rentStructure}
}

func (u *UtxoDB) genesisInit() {
	genesisTx, err := iotago.NewTransactionBuilder().
		AddInput(&iotago.ToBeSignedUTXOInput{Address: genesisAddress, Input: &iotago.UTXOInput{}}).
		AddOutput(&iotago.ExtendedOutput{Address: genesisAddress, Amount: DefaultIOTASupply}).
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

func (u *UtxoDB) mustGetFundsFromFaucetTx(target iotago.Address) *iotago.Transaction {
	unspent := u.getUTXOInputs(genesisAddress)
	if len(unspent) != 1 {
		panic("number of genesis outputs must be 1")
	}
	utxo := unspent[0]
	out := u.getOutput(utxo.ID()).(*iotago.ExtendedOutput)
	tx, err := iotago.NewTransactionBuilder().
		AddInput(&iotago.ToBeSignedUTXOInput{Address: genesisAddress, Input: utxo}).
		AddOutput(&iotago.ExtendedOutput{Address: target, Amount: FundsFromFaucetAmount}).
		AddOutput(&iotago.ExtendedOutput{Address: genesisAddress, Amount: out.Amount - FundsFromFaucetAmount}).
		Build(u.deSeriParams(), genesisSigner)
	if err != nil {
		panic(err)
	}
	return tx
}

// NewKeyPairByIndex deterministic private key
func (u *UtxoDB) NewKeyPairByIndex(index uint64) (cryptolib.KeyPair, *iotago.Ed25519Address) {
	var tmp8 [8]byte
	binary.LittleEndian.PutUint64(tmp8[:], index)
	h := hashing.HashData(u.seed[:], tmp8[:])
	keyPair := cryptolib.NewKeyPairFromSeed(cryptolib.Seed(h))
	addr := cryptolib.Ed25519AddressFromPubKey(keyPair.PublicKey)
	return keyPair, addr
}

// GetFundsFromFaucet sends FundsFromFaucetAmount IOTA tokens from the genesis address to the given address.
func (u *UtxoDB) GetFundsFromFaucet(target iotago.Address) (*iotago.Transaction, error) {
	tx := u.mustGetFundsFromFaucetTx(target)
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
				ConfUnix:    uint64(u.globalLogicalTime.Time.Unix()),
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
		panic(err)
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
func (u *UtxoDB) GetUnspentOutputs(addr iotago.Address) (iotago.Outputs, []*iotago.UTXOInput) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	return u.getUnspentOutputs(addr)
}

func (u *UtxoDB) getUnspentOutputs(addr iotago.Address) (iotago.Outputs, []*iotago.UTXOInput) {
	ret := iotago.Outputs{}
	ids := u.getUTXOInputs(addr)
	for _, input := range ids {
		out := u.getOutput(input.ID())
		if out == nil {
			panic("inconsistency: unspent output not found")
		}
		ret = append(ret, out)
	}
	return ret, ids
}

// GetAddressBalanceIotas returns the total amount of iotas owned by the address
func (u *UtxoDB) GetAddressBalanceIotas(addr iotago.Address) uint64 {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	ret := uint64(0)
	outputs, _ := u.getUnspentOutputs(addr)
	for _, out := range outputs {
		ret += out.Deposit()
	}
	return ret
}

// GetAddressBalances returns the total amount of iotas and tokens owned by the address
func (u *UtxoDB) GetAddressBalances(addr iotago.Address) *iscp.Assets {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	iotas := uint64(0)
	tokens := iotago.NativeTokenSum{}
	outputs, _ := u.getUnspentOutputs(addr)
	for _, out := range outputs {
		iotas += out.Deposit()
		if out, ok := out.(iotago.NativeTokenOutput); ok {
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
	}
	return iscp.AssetsFromNativeTokenSum(iotas, tokens)
}

// GetAliasOutputs collects all outputs of type iotago.AliasOutput for the address
func (u *UtxoDB) GetAliasOutputs(addr iotago.Address) ([]*iotago.AliasOutput, []*iotago.UTXOInput) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	outs, ids := u.getUnspentOutputs(addr)
	ret := make([]*iotago.AliasOutput, 0)
	retIds := make([]*iotago.UTXOInput, 0)
	for i, out := range outs {
		if o, ok := out.(*iotago.AliasOutput); ok {
			ret = append(ret, o)
			retIds = append(retIds, ids[i])
		}
	}
	return ret, retIds
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

func (u *UtxoDB) getUTXOInputs(addr iotago.Address) []*iotago.UTXOInput {
	var ret []*iotago.UTXOInput
	for oid := range u.utxo {
		out := u.getOutput(oid)
		utxoInput := oid.UTXOInput()
		if getOutputAddress(out, utxoInput).Equal(addr) {
			ret = append(ret, utxoInput)
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
