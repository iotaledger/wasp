package utxodb

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

const (
	DefaultIOTASupply = iotago.TokenSupply

	Mi = 1_000_000

	// RequestFundsAmount is how many iotas are returned from the faucet.
	RequestFundsAmount = 1 * Mi
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
	mutex                sync.RWMutex
	supply               uint64
	seed                 [cryptolib.SeedSize]byte
	rentStructure        *iotago.RentStructure
	milestones           []milestone
	milestoneIndexByTxID map[iotago.TransactionID]uint32
	utxo                 map[iotago.OutputID]*iotago.UTXOInput
}

type MilestoneInfo struct {
	Index     uint32
	Timestamp UnixSeconds
}

type milestone struct {
	timestamp UnixSeconds
	tx        *iotago.Transaction // for simplicity, only one transaction per milestone
}

type InitParams struct {
	timestamp     UnixSeconds
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
		timestamp:     0,
		supply:        DefaultIOTASupply,
		rentStructure: &iotago.RentStructure{},
		seed:          seedBytes,
	}
}

func WithTimestamp(timestamp UnixSeconds) *InitParams {
	return DefaultInitParams().WithTimestamp(timestamp)
}

func (i *InitParams) WithTimestamp(timestamp UnixSeconds) *InitParams {
	i.timestamp = timestamp
	return i
}

func WithRentStructure(r *iotago.RentStructure) *InitParams {
	return DefaultInitParams().WithRentStructure(r)
}

func (i *InitParams) WithRentStructure(r *iotago.RentStructure) *InitParams {
	i.rentStructure = r
	return i
}

func WithSupply(supply uint64) *InitParams {
	return DefaultInitParams().WithSupply(supply)
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
		supply:               p.supply,
		seed:                 p.seed,
		rentStructure:        p.rentStructure,
		milestoneIndexByTxID: make(map[iotago.TransactionID]uint32),
		utxo:                 make(map[iotago.OutputID]*iotago.UTXOInput),
	}
	u.genesisInit(p.timestamp)
	return u
}

func (u *UtxoDB) deSeriParams() *iotago.DeSerializationParameters {
	return &iotago.DeSerializationParameters{RentStructure: u.rentStructure}
}

func (u *UtxoDB) genesisInit(timestamp UnixSeconds) {
	genesisTx, err := iotago.NewTransactionBuilder().
		AddInput(&iotago.ToBeSignedUTXOInput{Address: genesisAddress, Input: &iotago.UTXOInput{}}).
		AddOutput(&iotago.ExtendedOutput{Address: genesisAddress, Amount: DefaultIOTASupply}).
		Build(u.deSeriParams(), genesisSigner)
	if err != nil {
		panic(err)
	}
	u.addTransaction(genesisTx, timestamp)
}

func (u *UtxoDB) addTransaction(tx *iotago.Transaction, timestamp UnixSeconds) {
	txID, err := tx.ID()
	if err != nil {
		panic(err)
	}
	u.milestones = append(u.milestones, milestone{
		timestamp: timestamp,
		tx:        tx,
	})
	u.milestoneIndexByTxID[*txID] = uint32(len(u.milestones) - 1)

	// delete consumed outputs from the ledger
	inputs, err := u.getTransactionInputs(tx)
	for outID := range inputs {
		delete(u.utxo, outID)
	}

	// add unspent outputs to the ledger
	for i := range tx.Essence.Outputs {
		utxo := &iotago.UTXOInput{
			TransactionID:          *txID,
			TransactionOutputIndex: uint16(i),
		}
		u.utxo[utxo.ID()] = utxo
	}

	u.checkLedgerBalance()
}

func (u *UtxoDB) latestMilestone() milestone {
	return u.milestones[u.milestoneIndex()]
}

func (u *UtxoDB) milestoneIndex() uint32 {
	return uint32(len(u.milestones)) - 1
}

func (u *UtxoDB) GenesisTransaction() *iotago.Transaction {
	return u.milestones[0].tx
}

func (u *UtxoDB) GenesisTransactionID() iotago.TransactionID {
	txID, err := u.GenesisTransaction().ID()
	if err != nil {
		panic(err)
	}
	return *txID
}

// GenesisPrivateKey returns the private key of the creator of genesis.
func (u *UtxoDB) GenesisPrivateKey() *cryptolib.PrivateKey {
	return &genesisKeyPair.PrivateKey
}

// GenesisAddress returns the genesis address.
func (u *UtxoDB) GenesisAddress() iotago.Address {
	return genesisAddress
}

func (u *UtxoDB) mustRequestFundsTx(target iotago.Address) *iotago.Transaction {
	unspent := u.getUTXOInputs(genesisAddress)
	if len(unspent) != 1 {
		panic("number of genesis outputs must be 1")
	}
	utxo := unspent[0]
	out := u.getOutput(utxo.ID()).(*iotago.ExtendedOutput)
	tx, err := iotago.NewTransactionBuilder().
		AddInput(&iotago.ToBeSignedUTXOInput{Address: genesisAddress, Input: utxo}).
		AddOutput(&iotago.ExtendedOutput{Address: target, Amount: RequestFundsAmount}).
		AddOutput(&iotago.ExtendedOutput{Address: genesisAddress, Amount: out.Amount - RequestFundsAmount}).
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

// RequestFunds sends RequestFundsAmount IOTA tokens from the genesis address to the given address.
func (u *UtxoDB) RequestFunds(target iotago.Address) (*iotago.Transaction, error) {
	tx := u.mustRequestFundsTx(target)
	return tx, u.AddTransaction(tx)
}

// Supply returns supply of the instance.
func (u *UtxoDB) Supply() uint64 {
	return u.supply
}

// IsConfirmed checks if the transaction is in the UtxoDB ledger.
func (u *UtxoDB) IsConfirmed(txID *iotago.TransactionID) bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	_, ok := u.milestoneIndexByTxID[*txID]
	return ok
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
				return nil, fmt.Errorf("output not found")
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
		return err
	}
	for outID := range inputs {
		if u.utxo[outID] == nil {
			return fmt.Errorf("referenced output is not unspent")
		}
	}

	{
		semValCtx := &iotago.SemanticValidationContext{
			ExtParas: &iotago.ExternalUnlockParameters{
				ConfMsIndex: u.milestoneIndex(),
				ConfUnix:    uint64(u.latestMilestone().timestamp),
			},
		}
		if err := tx.SemanticallyValidate(semValCtx, inputs); err != nil {
			return err
		}
	}

	return nil
}

// AddTransaction adds a transaction to UtxoDB, ensuring consistency of the UtxoDB ledger.
func (u *UtxoDB) AddTransaction(tx *iotago.Transaction, timestamp ...UnixSeconds) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	t := u.latestMilestone().timestamp + 1
	if len(timestamp) > 0 {
		if timestamp[0] < t {
			panic("timestamp must be >= latest milestone timestamp")
		}
		t = timestamp[0]
	}

	if err := u.validateTransaction(tx); err != nil {
		return err
	}

	u.addTransaction(tx, t)
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

// GetTransactionMilestoneInfo returns the milestone index and timestamp of the transaction
func (u *UtxoDB) GetTransactionMilestoneInfo(txID iotago.TransactionID) (MilestoneInfo, bool) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	idx, ok := u.milestoneIndexByTxID[txID]
	if !ok {
		return MilestoneInfo{}, false
	}
	return MilestoneInfo{Index: idx, Timestamp: u.milestones[idx].timestamp}, true
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
	return iscp.NewAssetsFromNativeTokenSum(iotas, tokens)
}

// GetAliasOutputs collects all outputs of type iotago.AliasOutput for the transaction.
func (u *UtxoDB) GetAliasOutputs(addr iotago.Address) []*iotago.AliasOutput {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	outs, _ := u.getUnspentOutputs(addr)
	ret := make([]*iotago.AliasOutput, 0)
	for _, out := range outs {
		if o, ok := out.(*iotago.AliasOutput); ok {
			ret = append(ret, o)
		}
	}
	return ret
}

func (u *UtxoDB) getTransaction(txID iotago.TransactionID) (*iotago.Transaction, bool) {
	milestoneIndex, ok := u.milestoneIndexByTxID[txID]
	if !ok {
		return nil, false
	}
	return u.milestones[milestoneIndex].tx, true
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
		// FIXME this is temporary patch of the bug in the iota.go AliasOutput.Chain() method
		aliasID := out.Chain().(iotago.AliasID)
		var nilAliasID iotago.AliasID
		if aliasID == nilAliasID {
			aliasID = iotago.AliasIDFromOutputID(id.ID())
		}
		return aliasID.ToAddress()
		// -- end FIXME
	default:
		panic("unknown ident output type")
	}
}

func (u *UtxoDB) getUTXOInputs(addr iotago.Address) []*iotago.UTXOInput {
	var ret []*iotago.UTXOInput
	for _, input := range u.utxo {
		out := u.getOutput(input.ID())
		if getOutputAddress(out, input).Equal(addr) {
			ret = append(ret, input)
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
