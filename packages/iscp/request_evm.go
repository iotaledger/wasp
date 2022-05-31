package iscp

import (
	"math"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

// EVMGasBookkeeping is the exceess iotas needed for gas when sending an EVM transaction
// TODO: automatically calculate this? Do not charge for blockchain bookkeeping?
const EVMGasBookkeeping = 100_000

func NewEVMOffLedgerRequest(chainID *ChainID, tx *types.Transaction, gasRatio *util.Ratio32) (*OffLedgerRequestData, error) {
	// TODO: verify tx.ChainId()?
	signatureScheme, err := newEVMOffLedferSignatureSchemeFromTransaction(tx)
	if err != nil {
		return nil, err
	}
	return &OffLedgerRequestData{
		chainID:         chainID,
		contract:        Hn(evmnames.Contract),
		entryPoint:      Hn(evmnames.FuncSendTransaction),
		params:          dict.Dict{evmnames.FieldTransaction: evmtypes.EncodeTransaction(tx)},
		signatureScheme: signatureScheme,
		nonce:           tx.Nonce(),
		allowance:       NewEmptyAllowance(),
		gasBudget:       evmtypes.EVMGasToISC(tx.Gas(), gasRatio) + EVMGasBookkeeping,
	}, nil
}

func NewEVMOffLedgerEstimateGasRequest(chainID *ChainID, callMsg ethereum.CallMsg, gasRatio *util.Ratio32) *OffLedgerRequestData {
	return &OffLedgerRequestData{
		chainID:         chainID,
		contract:        Hn(evmnames.Contract),
		entryPoint:      Hn(evmnames.FuncEstimateGas),
		params:          dict.Dict{evmnames.FieldCallMsg: evmtypes.EncodeCallMsg(callMsg)},
		signatureScheme: newEVMOffLedgerSignatureScheme(callMsg.From),
		nonce:           0,
		allowance:       NewEmptyAllowance(),
		gasBudget:       math.MaxUint64,
	}
}

func (r *OffLedgerRequestData) IsEVM() bool {
	return r.IsEVMSendTransaction() || r.IsEVMEstimateGas()
}

func (r *OffLedgerRequestData) IsEVMSendTransaction() bool {
	return r.contract == Hn(evmnames.Contract) && r.entryPoint == Hn(evmnames.FuncSendTransaction)
}

func (r *OffLedgerRequestData) IsEVMEstimateGas() bool {
	return r.contract == Hn(evmnames.Contract) && r.entryPoint == Hn(evmnames.FuncEstimateGas)
}

type evmOffLedgerSignatureScheme struct {
	sender *EthereumAddressAgentID
}

var _ OffLedgerSignatureScheme = &evmOffLedgerSignatureScheme{}

func newEVMOffLedferSignatureSchemeFromTransaction(tx *types.Transaction) (*evmOffLedgerSignatureScheme, error) {
	sender, err := evmutil.GetSender(tx)
	if err != nil {
		return nil, err
	}
	return newEVMOffLedgerSignatureScheme(sender), nil
}

func newEVMOffLedgerSignatureScheme(sender common.Address) *evmOffLedgerSignatureScheme {
	return &evmOffLedgerSignatureScheme{sender: NewEthereumAddressAgentID(sender)}
}

func (s *evmOffLedgerSignatureScheme) Sender() AgentID {
	return s.sender
}

func (*evmOffLedgerSignatureScheme) readEssence(mu *marshalutil.MarshalUtil) error {
	return nil
}

func (*evmOffLedgerSignatureScheme) readSignature(mu *marshalutil.MarshalUtil) error {
	return nil
}

func (*evmOffLedgerSignatureScheme) setPublicKey(key *cryptolib.PublicKey) {
	panic("should not be called")
}

func (*evmOffLedgerSignatureScheme) sign(key *cryptolib.KeyPair, data []byte) {
	panic("should not be called")
}

func (*evmOffLedgerSignatureScheme) verify(data []byte) bool {
	return true
}

func (*evmOffLedgerSignatureScheme) writeEssence(mu *marshalutil.MarshalUtil) {
}

func (*evmOffLedgerSignatureScheme) writeSignature(mu *marshalutil.MarshalUtil) {
}

var _ OffLedgerSignatureScheme = &evmOffLedgerSignatureScheme{}
