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
	extraData map[string][]byte,
) ([]byte, error) {
	chunkSize := 180

	// If payload is a single buffer, wrap it in a slice
	if len(payload) == 1 {
		payload = [][]byte{payload[0]}
	}

	parameterList := make([][]byte, 0)
	data := make(map[string][]byte)

	// Copy extraData to data
	for k, v := range extraData {
		data[k] = v
	}

	// Process each payload
	for _, currentPayload := range payload {
		chunkList := make([][]byte, 0)

		// Split payload into chunks
		for i := 0; i < len(currentPayload); i += chunkSize {
			end := i + chunkSize
			if end > len(currentPayload) {
				end = len(currentPayload)
			}
			chunk := currentPayload[i:end]
			chunkList = append(chunkList, chunk)
		}

		// Initialize lastHash
		lastHash := make([]byte, 32)
		l.log(lastHash)

		// Process chunks in reverse order
		for i := len(chunkList) - 1; i >= 0; i-- {
			chunk := chunkList[i]
			linkedChunk := bytes.Join([][]byte{lastHash, chunk}, []byte{})

			l.log("Chunk: ", chunk)
			l.log("linkedChunk: ", linkedChunk)

			// Calculate new hash
			hash := sha256.Sum256(linkedChunk)
			lastHash = hash[:]

			// Store in data map
			data[hex.EncodeToString(lastHash)] = linkedChunk
		}

		parameterList = append(parameterList, lastHash)
		lastHash = make([]byte, 32)
	}

	l.log(data)

	// Prepare final parameter buffer
	startBuf := []byte{byte(START)}
	finalBuf := append(startBuf, bytes.Join(parameterList, []byte{})...)

	return l.handleBlocksProtocol(cla, ins, p1, p2, finalBuf, data)
}
func (h *HWLedger) handleBlocksProtocol(
	cla, ins, p1, p2 uint8,
	initialPayload []byte,
	data map[string][]byte,
) ([]byte, error) {
	payload := initialPayload
	result := make([]byte, 0)

	for {
		h.log("Sending payload to ledger: ", hex.EncodeToString(payload), payload)

		// Construct request
		req := []byte{cla, ins, p1, p2, byte(len(payload))}
		req = append(req, payload...)

		h.log("formatted full request ", req, len(req))

		// Send request
		response, err := h.device.Exchange(req)
		if err != nil {
			return nil, fmt.Errorf("exchange error: %w", err)
		}
		h.log("Received response: ", hex.EncodeToString(response))

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
			h.log("Getting block ", hex.EncodeToString(rvPayload))
			chunk, exists := data[hex.EncodeToString(rvPayload)]
			h.log("Found block ", hex.EncodeToString(chunk))

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
