package registry

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"io"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/mr-tron/base58"
)

// ChainRecord is a minimum data needed to load a committee for the chain
// it is up to the node (not smart contract) to check authorizations to create/update this record
type ChainRecord struct {
	ChainID        coretypes.ChainID
	CommitteeNodes []string // "host_addr:port"
	Active         bool
}

func dbkeyChainRecord(chainID *coretypes.ChainID) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeChainRecord, chainID.Bytes())
}

func SaveChainRecord(bd *ChainRecord) error {
	if bd.ChainID == coretypes.NilChainID {
		return fmt.Errorf("can be empty chain id")
	}
	var buf bytes.Buffer
	if err := bd.Write(&buf); err != nil {
		return err
	}
	if err := database.GetRegistryPartition().Set(dbkeyChainRecord(&bd.ChainID), buf.Bytes()); err != nil {
		return err
	}
	publisher.Publish("chainrec", bd.ChainID.String(), bd.ChainID.String())
	return nil
}

func GetChainRecord(chainID *coretypes.ChainID) (*ChainRecord, error) {
	data, err := database.GetRegistryPartition().Get(dbkeyChainRecord(chainID))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ret := new(ChainRecord)
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func UpdateChainRecord(chainID *coretypes.ChainID, f func(*ChainRecord) bool) (*ChainRecord, error) {
	bd, err := GetChainRecord(chainID)
	if err != nil {
		return nil, err
	}
	if bd == nil {
		return nil, fmt.Errorf("no chain record found for address %s", chainID.String())
	}
	if f(bd) {
		err = SaveChainRecord(bd)
		if err != nil {
			return nil, err
		}
	}
	return bd, nil
}

func ActivateChainRecord(chainID *coretypes.ChainID) (*ChainRecord, error) {
	return UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func DeactivateChainRecord(chainID *coretypes.ChainID) (*ChainRecord, error) {
	return UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if !bd.Active {
			return false
		}
		bd.Active = false
		return true
	})
}

func GetChainRecords() ([]*ChainRecord, error) {
	db := database.GetRegistryPartition()
	ret := make([]*ChainRecord, 0)

	err := db.Iterate([]byte{dbprovider.ObjectTypeChainRecord}, func(key kvstore.Key, value kvstore.Value) bool {
		bd := new(ChainRecord)
		if err := bd.Read(bytes.NewReader(value)); err == nil {
			ret = append(ret, bd)
		} else {
			log.Warnf("corrupted chain record with key %s", base58.Encode(key))
		}
		return true
	})
	return ret, err
}

func (bd *ChainRecord) Write(w io.Writer) error {
	if err := bd.ChainID.Write(w); err != nil {
		return err
	}
	if err := util.WriteStrings16(w, bd.CommitteeNodes); err != nil {
		return err
	}
	if err := util.WriteBoolByte(w, bd.Active); err != nil {
		return err
	}
	return nil
}

func (bd *ChainRecord) Read(r io.Reader) error {
	var err error
	if err = bd.ChainID.Read(r); err != nil {
		return err
	}
	if bd.CommitteeNodes, err = util.ReadStrings16(r); err != nil {
		return err
	}
	if err = util.ReadBoolByte(r, &bd.Active); err != nil {
		return err
	}
	return nil
}

func (bd *ChainRecord) String() string {
	ret := "ChainID: " + bd.ChainID.String() + "\n"
	ret += fmt.Sprintf("      Committee nodes: %+v\n", bd.CommitteeNodes)
	return ret
}
