package registry

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/publisher"
	"io"
)

// each program is uniquely identified by the hash of it's code

type ProgramMetadata struct {
	// program hash. Persist in key
	ProgramHash hashing.HashValue
	// it is interpreted by the loader to locate and cache program's code
	Location string
	// description any text
	Description string
}

func dbkeyProgramMetadata(progHash *hashing.HashValue) []byte {
	return database.MakeKey(database.ObjectTypeProgramMetadata, progHash[:])
}

func dbkeyProgramCode(progHash *hashing.HashValue) []byte {
	return database.MakeKey(database.ObjectTypeProgramCode, progHash[:])
}

func GetProgramCode(progHash *hashing.HashValue) ([]byte, bool, error) {
	db := database.GetRegistryPartition()
	data, err := db.Get(dbkeyProgramCode(progHash))
	if err == kvstore.ErrKeyNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	hashData := hashing.HashData(data)
	if *progHash != *hashData {
		return nil, false, fmt.Errorf("program code is corrupted. Program hash: %s", progHash.String())
	}
	return data, true, nil
}

func SaveProgramCode(programCode []byte) (ret hashing.HashValue, err error) {
	progHash := hashing.HashData(programCode)
	db := database.GetRegistryPartition()
	if err = db.Set(dbkeyProgramCode(progHash), programCode); err != nil {
		return
	}
	ret = *progHash

	defer publisher.Publish("programcode", progHash.String())
	return
}

func SaveProgramMetadata(md *ProgramMetadata) error {
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

func GetProgramMetadata(progHash *hashing.HashValue) (*ProgramMetadata, bool, error) {
	db := database.GetRegistryPartition()
	data, err := db.Get(dbkeyProgramMetadata(progHash))
	if err == kvstore.ErrKeyNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	ret := &ProgramMetadata{}
	err = ret.Read(bytes.NewReader(data))
	if err != nil {
		return nil, false, err
	}
	ret.ProgramHash = *progHash
	return ret, true, nil
}

func (md *ProgramMetadata) Write(w io.Writer) error {
	if err := util.WriteString16(w, md.Location); err != nil {
		return err
	}
	if err := util.WriteString16(w, md.Description); err != nil {
		return err
	}
	return nil
}

func (md *ProgramMetadata) Read(r io.Reader) error {
	var err error
	if md.Location, err = util.ReadString16(r); err != nil {
		return err
	}
	if md.Description, err = util.ReadString16(r); err != nil {
		return err
	}
	return nil
}
