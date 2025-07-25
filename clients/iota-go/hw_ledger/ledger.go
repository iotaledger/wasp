package hw_ledger

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	ledgergo "github.com/iotaledger/wasp/v2/clients/iota-go/hw_ledger/ledger-go"
)

var (
	ErrDeviceNotFound          = errors.New("device not found")
	ErrConnectionFailure       = errors.New("failed to connect to device")
	ErrInvalidResponseLength   = errors.New("invalid response length")
	ErrTooManyLedgersConnected = errors.New("too many ledgers connected")
)

type HWLedger struct {
	device ledgergo.LedgerDevice
}

func NewHWLedger(device ledgergo.LedgerDevice) *HWLedger {
	return &HWLedger{
		device: device,
	}
}

// TryAndConnect tries to connect to a Ledger (index 0). Multiple ledgers are not supported.
func TryAndConnect() (*HWLedger, error) {
	l := ledgergo.NewLedgerHIDTransport()
	count := l.CountDevices()

	if count == 0 {
		return nil, ErrDeviceNotFound
	}

	if count > 1 {
		return nil, ErrTooManyLedgersConnected
	}

	device, err := l.Connect(0)
	if err != nil {
		return nil, errors.Join(ErrConnectionFailure, err)
	}

	return &HWLedger{device: device}, nil
}

func (l *HWLedger) Close() error {
	return l.device.Close()
}

func (l *HWLedger) log(args ...interface{}) {
	fmt.Println(args...)
}

func (l *HWLedger) GetVersion() (VersionResult, error) {
	const (
		cla = uint8(0x00)
		ins = uint8(0x00)
		p1  = uint8(0x00)
		p2  = uint8(0x00)
	)

	result, err := l.sendChunks(cla, ins, p1, p2, [][]byte{{0x0}})
	if err != nil {
		return VersionResult{}, err
	}

	if len(result) != VersionExpectedSize {
		return VersionResult{}, ErrInvalidResponseLength
	}

	return VersionResult{
		Major: result[0],
		Minor: result[1],
		Patch: result[2],
		Name:  string(result[3:]),
	}, nil
}

func (l *HWLedger) GetPublicKey(path string, displayOnDevice bool) (PublicKeyResult, error) {
	// Determine instruction based on displayOnDevice
	const (
		cla = uint8(0x00)
		p1  = uint8(0)
		p2  = uint8(0)
	)

	ins := uint8(0x02)
	if displayOnDevice {
		ins = 0x01
	}

	// Build BIP32 key payload
	payload, err := buildBip32KeyPayload(path)
	if err != nil {
		return PublicKeyResult{}, fmt.Errorf("failed to build payload: %w", err)
	}

	// Send chunks to get public key
	response, err := l.sendChunks(cla, ins, p1, p2, [][]byte{payload})
	if err != nil {
		return PublicKeyResult{}, fmt.Errorf("failed to send chunks: %w", err)
	}

	if len(response) != PublicKeyExpectedSize+2 { // +2 for the length of each item
		return PublicKeyResult{}, ErrInvalidResponseLength
	}

	const keySize = 32

	if response[0] != keySize || response[33] != keySize {
		return PublicKeyResult{}, ErrInvalidResponseLength
	}

	publicKey := response[1 : keySize+1]
	address := response[keySize+2 : keySize+2+keySize]

	// Return public key result
	return PublicKeyResult{
		PublicKey: [32]byte(publicKey),
		Address:   [32]byte(address),
	}, nil
}

func (l *HWLedger) SignTransaction(path string, txDataBytes []byte) (*SignTransactionResult, error) {
	const (
		cla = uint8(0x00)
		ins = uint8(0x03)
		p1  = uint8(0x00)
		p2  = uint8(0x00)
	)

	// Create hash size buffer (uint32 little-endian)
	hashSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(hashSize, uint32(len(txDataBytes)))

	bip32KeyPayload, err := buildBip32KeyPayload(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build BIP32 key payload: %w", err)
	}

	// Combine hash size and raw transaction
	payloadTxn := slices.Concat(hashSize, txDataBytes)
	l.log("Payload Txn", payloadTxn)

	// Send chunks and get signature
	signature, err := l.sendChunks(cla, ins, p1, p2, [][]byte{payloadTxn, bip32KeyPayload})
	if err != nil {
		return nil, fmt.Errorf("failed to send chunks: %w", err)
	}

	if len(signature) != SignTransactionExpectedSize {
		return nil, ErrInvalidResponseLength
	}

	return &SignTransactionResult{
		Signature: [64]byte(signature),
	}, nil
}

func buildBip32KeyPayload(path string) ([]byte, error) {
	paths, err := splitPath(path)
	if err != nil {
		return nil, err
	}

	// Allocate buffer: 1 byte for length + 4 bytes per path element
	payload := make([]byte, 1+len(paths)*4)
	payload[0] = byte(len(paths))

	// Write each path element as little-endian uint32
	for i, element := range paths {
		binary.LittleEndian.PutUint32(payload[1+4*i:], uint32(element))
	}

	return payload, nil
}

// splitPath splits a supplied bip32 path into separate path elements
func splitPath(path string) ([]uint32, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}

	// Remove leading '/' if present
	if path[0] == '/' {
		path = path[1:]
	}

	components := strings.Split(path, "/")
	result := make([]uint32, 0, len(components))

	for _, element := range components {
		if element == "" {
			continue
		}

		// Check if hardened (ends with ')
		hardened := false
		if strings.HasSuffix(element, "'") {
			hardened = true
			element = element[:len(element)-1]
		}

		// Parse number
		num, err := strconv.ParseUint(element, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid path element '%s': %w", element, err)
		}

		if hardened {
			num += 0x80000000
		}

		result = append(result, uint32(num))
	}

	if len(result) == 0 {
		return nil, errors.New("no valid path elements")
	}

	return result, nil
}
