package registry

import (
	"bytes"
	"fmt"
	"io"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/packages/publisher"
)

// each program is uniquely identified by the hash of its binary code

type ProgramMetadata struct {
	// program hash. Persist in key
	ProgramHash hashing.HashValue
	// VM type. It is used to distinguish between several types of VMs
	VMType string
	// description any text
	Description string
}

var builtinPrograms = make(map[hashing.HashValue]*ProgramMetadata)

func RegisterBuiltinProgramMetadata(progHash *hashing.HashValue, description string) {
	builtinPrograms[*progHash] = &ProgramMetadata{
		ProgramHash: *progHash,
		VMType:      "builtin",
		Description: description,
	}
}

func dbkeyProgramMetadata(progHash *hashing.HashValue) []byte {
	return database.MakeKey(database.ObjectTypeProgramMetadata, progHash[:])
}

func (md *ProgramMetadata) Save() error {
	_, ok := builtinPrograms[md.ProgramHash]
	if ok {
		return fmt.Errorf("Cannot save builtin program %s", md.ProgramHash.String())
	}

	data, err := util.Bytes(md)
	if err != nil {
		return err
	}

	db := database.GetRegistryPartition()
	if err = db.Set(dbkeyProgramMetadata(&md.ProgramHash), data); err != nil {
		return err
	}

	defer publisher.Publish("programmetadata", md.ProgramHash.String())
	return nil
}

func GetProgramMetadata(progHash *hashing.HashValue) (*ProgramMetadata, error) {
	md, ok := builtinPrograms[*progHash]
	if ok {
		return md, nil
	}

	db := database.GetRegistryPartition()
	data, err := db.Get(dbkeyProgramMetadata(progHash))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ret := &ProgramMetadata{}
	err = ret.Read(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	ret.ProgramHash = *progHash
	return ret, nil
}

func (md *ProgramMetadata) Write(w io.Writer) error {
	if err := util.WriteString16(w, md.VMType); err != nil {
		return err
	}
	if err := util.WriteString16(w, md.Description); err != nil {
		return err
	}
	return nil
}

func (md *ProgramMetadata) Read(r io.Reader) error {
	var err error
	if md.VMType, err = util.ReadString16(r); err != nil {
		return err
	}
	if md.Description, err = util.ReadString16(r); err != nil {
		return err
	}
	return nil
}
