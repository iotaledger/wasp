package hw_ledger

// LedgerToHost represents commands sent from Ledger to Host
type LedgerToHost byte

const (
	ResultAccumulating LedgerToHost = iota // 0
	ResultFinal                            // 1
	GetChunk                               // 2
	PutChunk                               // 3
)

// String provides a string representation of LedgerToHost
func (l LedgerToHost) String() string {
	switch l {
	case ResultAccumulating:
		return "RESULT_ACCUMULATING"
	case ResultFinal:
		return "RESULT_FINAL"
	case GetChunk:
		return "GET_CHUNK"
	case PutChunk:
		return "PUT_CHUNK"
	default:
		return "UNKNOWN"
	}
}

// HostToLedger represents commands sent from Host to Ledger
type HostToLedger byte

const (
	START                      HostToLedger = iota // 0
	GetChunkResponseSuccess                        // 1
	GetChunkResponseFailure                        // 2
	PutChunkResponse                               // 3
	ResultAccumulatingResponse                     // 4
)

// String provides a string representation of HostToLedger
func (h HostToLedger) String() string {
	switch h {
	case START:
		return "START"
	case GetChunkResponseSuccess:
		return "GET_CHUNK_RESPONSE_SUCCESS"
	case GetChunkResponseFailure:
		return "GET_CHUNK_RESPONSE_FAILURE"
	case PutChunkResponse:
		return "PUT_CHUNK_RESPONSE"
	case ResultAccumulatingResponse:
		return "RESULT_ACCUMULATING_RESPONSE"
	default:
		return "UNKNOWN"
	}
}

type VersionResult struct {
	Major byte
	Minor byte
	Patch byte
	Name  string
}

type PublicKeyResult struct {
	PublicKey []byte
	Address   []byte
}

type SignTransactionResult struct {
	Signature []byte
}
