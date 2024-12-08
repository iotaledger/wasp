package hw_ledger

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	ledgergo "github.com/iotaledger/wasp/clients/iota-go/hw_ledger/ledger-go"
)

var (
	ErrDeviceNotFound    = errors.New("device not found")
	ErrConnectionFailure = errors.New("failed to connect to device")
)

type HWLedger struct {
	device ledgergo.LedgerDevice
}

func NewHWLedger(device ledgergo.LedgerDevice) *HWLedger {
	return &HWLedger{
		device: device,
	}
}

func TryAndConnect() (*HWLedger, error) {
	l := ledgergo.NewLedgerAdmin()

	if l.CountDevices() == 0 {
		return nil, ErrDeviceNotFound
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
	result, err := l.sendChunks(
		0x00, 0x00, 0x00, 0x00, [][]byte{
			{0x0},
		}, nil,
	)
	if err != nil {
		return VersionResult{}, err
	}

	return VersionResult{
		Major: result[0],
		Minor: result[1],
		Patch: result[2],
		Name:  string(result[3:]),
	}, nil
}

func (h *HWLedger) GetPublicKey(path string, displayOnDevice bool) (*PublicKeyResult, error) {
	cla := uint8(0x00)
	var ins uint8
	if displayOnDevice {
		ins = 0x01
	} else {
		ins = 0x02
	}
	p1 := uint8(0)
	p2 := uint8(0)

	payload, err := buildBip32KeyPayload(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build payload: %w", err)
	}

	response, err := h.sendChunks(
		cla, ins, p1, p2, [][]byte{
			payload,
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send chunks: %w", err)
	}

	if len(response) < 1 {
		return nil, errors.New("response too short")
	}

	keySize := response[0]
	if len(response) < int(keySize)+1 {
		return nil, errors.New("response shorter than key size")
	}

	publicKey := response[1 : keySize+1]
	var address []byte

	if len(response) > int(keySize)+2 {
		addressSize := response[keySize+1]
		if len(response) >= int(keySize)+2+int(addressSize) {
			address = response[keySize+2 : keySize+2+addressSize]
		}
	}

	return &PublicKeyResult{
		PublicKey: publicKey,
		Address:   address,
	}, nil
}

func (h *HWLedger) SignTransaction(path string, txDataBytes []byte) (*SignTransactionResult, error) {
	const (
		cla = uint8(0x00)
		ins = uint8(0x03)
		p1  = uint8(0)
		p2  = uint8(0)
	)

	// Create hash size buffer (uint32 little-endian)
	hashSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(hashSize, uint32(len(txDataBytes)))

	// Get BIP32 key payload
	bip32KeyPayload, err := buildBip32KeyPayload(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build BIP32 key payload: %w", err)
	}

	// Combine hash size and raw transaction
	payloadTxn := append(hashSize, txDataBytes...)
	h.log("Payload Txn", payloadTxn)

	// Send chunks and get signature
	signature, err := h.sendChunks(cla, ins, p1, p2, [][]byte{payloadTxn, bip32KeyPayload}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send chunks: %w", err)
	}

	return &SignTransactionResult{
		Signature: signature,
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
