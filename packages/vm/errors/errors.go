package errors

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
)

type ErrorDefinition struct {
	id            uint16
	messageFormat string
}

func (e *ErrorDefinition) Id() uint16 {
	return e.id
}

func (e *ErrorDefinition) MessageFormat() string {
	return e.messageFormat
}

func (e *ErrorDefinition) Create(params ...interface{}) BlockError {
	return BlockError{
		Id:            e.Id(),
		MessageFormat: e.MessageFormat(),
		Params:        params,
	}
}

type BlockError struct {
	Id            uint16
	MessageFormat string
	Params        []interface{}
}

func (e *BlockError) Message() string {
	return fmt.Sprintf(e.MessageFormat, e.Params...)
}

func (e *BlockError) AsError() error {
	return xerrors.Errorf(e.MessageFormat, e.Params...)
}

func (e *BlockError) Hash() string {
	hashValue := hashing.HashData([]byte(e.Message()))

	return hashValue.Hex()
}

func (e *BlockError) serializeParams(mu *marshalutil.MarshalUtil) error {
	bytes, err := json.Marshal(e.Params)

	mu.WriteUint16(uint16(len(bytes)))
	mu.WriteBytes(bytes)

	return err
}

func (e *BlockError) Serialize(mu *marshalutil.MarshalUtil) error {
	hash := []byte(e.Hash())

	mu.WriteUint16(e.Id).
		WriteUint16(uint16(len(hash))).
		WriteBytes(hash)

	// For now, JSON encoded.
	err := e.serializeParams(mu)

	return err
}

func NewBlockErrorDefinition(id uint16, messageFormat string) ErrorDefinition {
	return ErrorDefinition{id: id, messageFormat: messageFormat}
}

func (e *BlockError) deserializeParams(mu *marshalutil.MarshalUtil) error {
	var err error
	var paramLength uint16
	var params []byte

	if paramLength, err = mu.ReadUint16(); err != nil {
		return err
	}

	if params, err = mu.ReadBytes(int(paramLength)); err != nil {
		return err
	}

	if err = json.Unmarshal(params, &e.Params); err != nil {
		return err
	}

	return err
}

func ErrorFromBytes(mu *marshalutil.MarshalUtil, e ErrorCollection) (*BlockError, error) {
	var err error
	var hash string
	var hashLength uint16
	var errorId uint16
	var blockError *BlockError

	if errorId, err = mu.ReadUint16(); err != nil {
		return nil, err
	}

	if errorId == 0 {
		return nil, nil
	}

	if blockError, err = e.Create(int(errorId)); err != nil {
		return nil, err
	}

	if hashLength, err = mu.ReadUint16(); err != nil {
		return nil, err
	}

	if hashInBytes, err := mu.ReadBytes(int(hashLength)); err != nil {
		return nil, err
	} else {
		hash = string(hashInBytes)
	}

	if err = blockError.deserializeParams(mu); err != nil {
		return nil, err
	}

	newHash := blockError.Hash()

	if newHash != hash {
		return nil, xerrors.Errorf("Hash of supplied error does not match the serialized form! %v:%v", hash, newHash)
	}

	return blockError, nil
}
