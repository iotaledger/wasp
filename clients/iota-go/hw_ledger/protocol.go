package hw_ledger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

func (l *HWLedger) sendChunks(
	cla uint8,
	ins uint8,
	p1 uint8,
	p2 uint8,
	payload [][]byte,
) ([]byte, error) {
	// Define chunk size
	const chunkSize = 180

	// Initialize data map and parameter list
	data := make(map[string][]byte)
	parameterList := make([][]byte, 0)

	// Process each payload
	for _, currentPayload := range payload {
		// Split payload into chunks
		chunkList := splitPayload(currentPayload, chunkSize)

		// Initialize lastHash
		lastHash := make([]byte, 32)

		// Process chunks in reverse order
		for i := len(chunkList) - 1; i >= 0; i-- {
			chunk := chunkList[i]
			linkedChunk := bytes.Join([][]byte{lastHash, chunk}, []byte{})

			// Calculate new hash
			hash := sha256.Sum256(linkedChunk)
			lastHash = hash[:]

			// Store in data map
			data[hex.EncodeToString(lastHash)] = linkedChunk
		}

		// Append lastHash to parameter list
		parameterList = append(parameterList, lastHash)
	}

	// Prepare final parameter buffer
	startBuf := []byte{byte(START)}
	finalBuf := append(startBuf, bytes.Join(parameterList, []byte{})...)

	// Call handleBlocksProtocol
	return l.handleBlocksProtocol(cla, ins, p1, p2, finalBuf, data)
}

// Define a function to split payload into chunks
func splitPayload(payload []byte, chunkSize int) [][]byte {
	chunkList := make([][]byte, 0)
	for i := 0; i < len(payload); i += chunkSize {
		end := i + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		chunk := payload[i:end]
		chunkList = append(chunkList, chunk)
	}
	return chunkList
}

// Define a function to handle blocks protocol
func (l *HWLedger) handleBlocksProtocol(
	cla, ins, p1, p2 uint8,
	initialPayload []byte,
	data map[string][]byte,
) ([]byte, error) {
	payload := initialPayload
	result := make([]byte, 0)

	for {
		// Construct request
		req := []byte{cla, ins, p1, p2, byte(len(payload))}
		req = append(req, payload...)

		// Send request
		response, err := l.device.Exchange(req)
		if err != nil {
			return nil, fmt.Errorf("exchange error: %w", err)
		}

		// Validate response
		if len(response) < 3 { // Need at least instruction byte and 2 status bytes
			return nil, errors.New("response too short")
		}

		rvInstruction := LedgerToHost(response[0])
		rvPayload := response[1 : len(response)-2] // Last two bytes are return code

		// Validate instruction
		if rvInstruction > PutChunk {
			return nil, errors.New("unknown instruction returned from ledger")
		}

		switch rvInstruction {
		case ResultAccumulating, ResultFinal:
			result = append(result, rvPayload...)
			payload = []byte{byte(ResultAccumulatingResponse)}

			if rvInstruction == ResultFinal {
				return result, nil
			}

		case GetChunk:
			chunk, exists := data[hex.EncodeToString(rvPayload)]
			if exists {
				payload = append([]byte{byte(GetChunkResponseSuccess)}, chunk...)
			} else {
				payload = []byte{byte(GetChunkResponseFailure)}
			}

		case PutChunk:
			hash := sha256.Sum256(rvPayload)
			data[hex.EncodeToString(hash[:])] = rvPayload
			payload = []byte{byte(PutChunkResponse)}
		}
	}
}
