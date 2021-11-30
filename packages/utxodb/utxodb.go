package utxodb

import (
	"fmt"
	"sync"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
)

const (
	DefaultIOTASupply = iotago.TokenSupply

	Mi = 1_000_000

	// RequestFundsAmount is how many iotas are returned from the faucet.
	RequestFundsAmount = 1 * Mi
)

var (
	genesisKey     = ed25519.NewKeyFromSeed([]byte("3.141592653589793238462643383279"))
	genesisAddress = iotago.Ed25519AddressFromPubKey(genesisKey.Public().(ed25519.PublicKey))
	genesisSigner  = iotago.NewInMemoryAddressSigner(iotago.NewAddressKeysForEd25519Address(&genesisAddress, genesisKey))
)

type UnixSeconds uint64

// UtxoDB mocks the Tangle ledger by implementing a fully synchronous in-memory database
// of transactions. It ensures the consistency of the ledger and all added transactions
// by checking inputs, outputs and signatures.
type UtxoDB struct {
	mutex  sync.RWMutex
	supply uint64
	// enqueuedTxs contains all transactions that are not yet committed in a milestone
	enqueuedTxs map[iotago.TransactionID]*iotago.Transaction
	// transactions contains all committed transactions
	transactions map[iotago.TransactionID]*iotago.Transaction
	milestones   []milestone
	utxo         map[iotago.OutputID]*iotago.UTXOInput
}

type milestone struct {
	timestamp    UnixSeconds
	transactions []iotago.TransactionID
}

type InitParams struct {
	timestamp UnixSeconds
	supply    uint64
}

func DefaultInitParams() *InitParams {
	return &InitParams{
		timestamp: 0,
		supply:    DefaultIOTASupply,
	}
}

func WithTimestamp(timestamp UnixSeconds) *InitParams {
	i := DefaultInitParams()
	i.timestamp = timestamp
	return i
}

func (i *InitParams) WithTimestamp(timestamp UnixSeconds) *InitParams {
	i.timestamp = timestamp
	return i
}

func WithSupply(supply uint64) *InitParams {
	i := DefaultInitParams()
	i.supply = supply
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
		enqueuedTxs:  make(map[iotago.TransactionID]*iotago.Transaction),
		transactions: make(map[iotago.TransactionID]*iotago.Transaction),
		utxo:         make(map[iotago.OutputID]*iotago.UTXOInput),
	}
	u.genesisInit(p.timestamp)
	return u
}

var deSeriParas = &iotago.DeSerializationParameters{RentStructure: &iotago.RentStructure{
	VByteCost:    0,
	VBFactorData: 0,
	VBFactorKey:  0,
}}

func (u *UtxoDB) genesisInit(timestamp UnixSeconds) {
	genesisTx, err := iotago.NewTransactionBuilder().
		AddInput(&iotago.ToBeSignedUTXOInput{Address: &genesisAddress, Input: &iotago.UTXOInput{}}).
		AddOutput(&iotago.ExtendedOutput{Address: &genesisAddress, Amount: DefaultIOTASupply}).
		Build(deSeriParas, genesisSigner)
	if err != nil {
		panic(err)
	}

	txID, err := genesisTx.ID()
	if err != nil {
		panic(err)
	}
	u.transactions[*txID] = genesisTx
	u.milestones = append(u.milestones, milestone{
		timestamp:    timestamp,
		transactions: []iotago.TransactionID{*txID},
	})
	utxo := &iotago.UTXOInput{
		TransactionID:          *txID,
		TransactionOutputIndex: 0,
	}
	u.utxo[utxo.ID()] = utxo
}

type CommitError struct {
	error
	TxID iotago.TransactionID
}

// Commit creates a new milestone with all transactions added since the latest milestone,
// returning the list of discarded transactions (e.g. because of double spend).
func (u *UtxoDB) Commit(timestamp ...UnixSeconds) (errors []*CommitError) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	t := u.latestMilestone().timestamp + 1
	if len(timestamp) > 0 {
		if timestamp[0] < t {
			panic("timestamp must be >= latest milestone timestamp")
		}
		t = timestamp[0]
	}

	// process transactions in a deterministic order
	txIDs := make([]iotago.TransactionID, 0, len(u.enqueuedTxs))
	for txID := range u.enqueuedTxs {
		txIDs = append(txIDs, txID)
	}
	txIDs = serializer.RemoveDupsAndSortByLexicalOrderArrayOf32Bytes(txIDs)

	for _, txID := range txIDs {
		tx := u.enqueuedTxs[txID]

		inputs, err := u.validateTransaction(tx)
		if err != nil {
			errors = append(errors, &CommitError{error: err, TxID: txID})
			continue
		}

		// delete consumed outputs from the ledger
		for outID := range inputs {
			delete(u.utxo, outID)
		}

		u.transactions[txID] = tx

		// add unspent outputs to the ledger
		for i := range tx.Essence.Outputs {
			utxo := &iotago.UTXOInput{
				TransactionID:          txID,
				TransactionOutputIndex: uint16(i),
			}
			u.utxo[utxo.ID()] = utxo
		}
	}

	u.milestones = append(u.milestones, milestone{
		timestamp:    t,
		transactions: txIDs,
	})
	u.enqueuedTxs = make(map[iotago.TransactionID]*iotago.Transaction)

	u.checkLedgerBalance()
	return
}

func (u *UtxoDB) latestMilestone() milestone {
	return u.milestones[u.milestoneIndex()]
}

func (u *UtxoDB) milestoneIndex() uint32 {
	return uint32(len(u.milestones)) - 1
}

func (u *UtxoDB) GenesisTransaction() *iotago.Transaction {
	return u.transactions[u.GenesisTransactionID()]
}

func (u *UtxoDB) GenesisTransactionID() iotago.TransactionID {
	return u.milestones[0].transactions[0]
}

// GenesisKey returns the private key of the creator of genesis.
func (u *UtxoDB) GenesisKey() *ed25519.PrivateKey {
	return &genesisKey
}

// GenesisAddress returns the genesis address.
func (u *UtxoDB) GenesisAddress() iotago.Address {
	return &genesisAddress
}

func (u *UtxoDB) mustRequestFundsTx(target iotago.Address) *iotago.Transaction {
	unspent := u.getUTXOInputs(&genesisAddress)
	if len(unspent) != 1 {
		panic("number of genesis outputs must be 1")
	}
	utxo := unspent[0]
	out := u.getOutput(utxo.ID()).(*iotago.ExtendedOutput)
	tx, err := iotago.NewTransactionBuilder().
		AddInput(&iotago.ToBeSignedUTXOInput{Address: &genesisAddress, Input: utxo}).
		AddOutput(&iotago.ExtendedOutput{Address: target, Amount: RequestFundsAmount}).
		AddOutput(&iotago.ExtendedOutput{Address: &genesisAddress, Amount: out.Amount - RequestFundsAmount}).
		Build(deSeriParas, genesisSigner)
	if err != nil {
		panic(err)
	}
	return tx
}

// RequestFunds sends RequestFundsAmount IOTA tokens from the genesis address to the given address.
func (u *UtxoDB) RequestFunds(target iotago.Address) (*iotago.Transaction, error) {
	tx := u.mustRequestFundsTx(target)
	return tx, u.AddTransactionAndCommit(tx)
}

// Supply returns supply of the instance.
func (u *UtxoDB) Supply() uint64 {
	return u.supply
}

// IsConfirmed checks if the transaction is in the UtxoDB ledger.
func (u *UtxoDB) IsConfirmed(txid *iotago.TransactionID) bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	_, ok := u.transactions[*txid]
	return ok
}

// getOutput finds an output by ID (either spent or unspent).
func (u *UtxoDB) GetOutput(outID iotago.OutputID) iotago.Output {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.getOutput(outID)
}

func (u *UtxoDB) getOutput(outID iotago.OutputID) iotago.Output {
	tx := u.transactions[outID.TransactionID()]
	if tx == nil {
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

func (u *UtxoDB) validateTransaction(tx *iotago.Transaction) (iotago.OutputSet, error) {
	// serialize for syntactic check
	if _, err := tx.Serialize(serializer.DeSeriModePerformValidation, deSeriParas); err != nil {
		return nil, err
	}

	inputs, err := u.getTransactionInputs(tx)
	if err != nil {
		return nil, err
	}
	for outID := range inputs {
		if u.utxo[outID] == nil {
			return nil, fmt.Errorf("referenced output is not unspent")
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
			return nil, err
		}
	}

	return inputs, nil
}

// AddTransaction adds a transaction to UtxoDB or returns an error.
// The function ensures consistency of the UtxoDB ledger.
// The transaction is unconfirmed until Commit is called.
func (u *UtxoDB) AddTransaction(tx *iotago.Transaction) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if _, err := u.validateTransaction(tx); err != nil {
		return err
	}

	txID, err := tx.ID()
	if err != nil {
		panic(err)
	}
	u.enqueuedTxs[*txID] = tx
	return nil
}

func (u *UtxoDB) AddTransactionAndCommit(tx *iotago.Transaction) error {
	err := u.AddTransaction(tx)
	if err != nil {
		return err
	}
	errors := u.Commit()
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// GetTransaction retrieves value transaction by its hash (ID).
func (u *UtxoDB) GetTransaction(id iotago.TransactionID) (*iotago.Transaction, bool) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	return u.getTransaction(id)
}

// MustGetTransaction same as GetTransaction only panics if transaction is not in UtxoDB.
func (u *UtxoDB) MustGetTransaction(id iotago.TransactionID) *iotago.Transaction {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.mustGetTransaction(id)
}

// GetUnspentOutputs returns all unspent outputs locked by the address, as spendable inputs.
func (u *UtxoDB) GetUnspentOutputs(addr iotago.Address) iotago.Outputs {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	ret := iotago.Outputs{}
	for _, input := range u.getUTXOInputs(addr) {
		out := u.getOutput(input.ID())
		if out == nil {
			panic("inconsistency: unspent output not found")
		}
		ret = append(ret, out)
	}
	return ret
}

// GetAddressBalance returns the total amount of iotas owned by the address
func (u *UtxoDB) GetAddressBalance(addr iotago.Address) uint64 {
	ret := uint64(0)
	outputs := u.GetUnspentOutputs(addr)
	for _, out := range outputs {
		ret += out.Deposit()
	}
	return ret
}

// GetAliasOutputs collects all outputs of type iotago.AliasOutput for the transaction.
func (u *UtxoDB) GetAliasOutputs(addr iotago.Address) []*iotago.AliasOutput {
	outs := u.GetUnspentOutputs(addr)
	ret := make([]*iotago.AliasOutput, 0)
	for _, out := range outs {
		if o, ok := out.(*iotago.AliasOutput); ok {
			ret = append(ret, o)
		}
	}
	return ret
}

func (u *UtxoDB) getTransaction(id iotago.TransactionID) (*iotago.Transaction, bool) {
	tx, ok := u.transactions[id]
	return tx, ok
}

func (u *UtxoDB) mustGetTransaction(id iotago.TransactionID) *iotago.Transaction {
	tx, ok := u.transactions[id]
	if !ok {
		panic(fmt.Errorf("utxodb.mustGetTransaction: tx id doesn't exist: %s", id))
	}
	return tx
}

func getOutputAddress(out iotago.Output) iotago.Address {
	switch out := out.(type) {
	case iotago.TransIndepIdentOutput:
		return out.Ident()
	case iotago.TransDepIdentOutput:
		return out.Chain().ToAddress()
	default:
		panic("unknown ident output type in semantic unlocks")
	}
}

func (u *UtxoDB) getUTXOInputs(addr iotago.Address) []*iotago.UTXOInput {
	var ret []*iotago.UTXOInput
	for _, input := range u.utxo {
		out := u.getOutput(input.ID())
		if getOutputAddress(out).Equal(addr) {
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
