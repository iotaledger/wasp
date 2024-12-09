package hw_ledger

type LedgerError int

var (
	LedgerErrorNotSupported = 0x6808
)

// LedgerToHost represents commands sent from Ledger to Host
type LedgerToHost byte

const (
	ResultAccumulating LedgerToHost = iota // 0
	ResultFinal                            // 1
	GetChunk                               // 2
	PutChunk                               // 3
)

// HostToLedger represents commands sent from Host to Ledger
type HostToLedger byte

const (
	START                      HostToLedger = iota // 0
	GetChunkResponseSuccess                        // 1
	GetChunkResponseFailure                        // 2
	PutChunkResponse                               // 3
	ResultAccumulatingResponse                     // 4
)

const VersionExpectedSize = 3 + 4 // 3 version bytes + 4 bytes as string (iota)
type VersionResult struct {
	Major byte
	Minor byte
	Patch byte
	Name  string
}

const PublicKeyExpectedSize = 32 + 32

type PublicKeyResult struct {
	PublicKey [32]byte
	Address   [32]byte
}

const SignTransactionExpectedSize = 64

type SignTransactionResult struct {
	Signature [64]byte
}
