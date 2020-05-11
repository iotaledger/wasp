package registry

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"go.dedis.ch/kyber/v3"
)

func dbKey(addr *address.Address) []byte {
	var buf bytes.Buffer
	buf.WriteString("key_")
	buf.Write(addr.Bytes())
	return buf.Bytes()
}

func CommitDKShare(ks *tcrypto.DKShare, pubKeys []kyber.Point) error {
	if err := ks.FinalizeDKS(pubKeys); err != nil {
		return err
	}
	return SaveDKShareToRegistry(ks)
}

func SaveDKShareToRegistry(ks *tcrypto.DKShare) error {
	if !ks.Committed {
		return fmt.Errorf("uncommited DK share: can't be saved to the registry")
	}
	dbase, err := database.GetDB()
	if err != nil {
		return err
	}
	dbkey := dbKey(ks.Address)
	exists, err := dbase.Contains(dbkey)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}

	var buf bytes.Buffer

	err = ks.Write(&buf)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbkey,
		Value: buf.Bytes(),
	})
}

func LoadDKShare(address *address.Address, maskPrivate bool) (*tcrypto.DKShare, error) {
	dbase, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	entry, err := dbase.Get(dbKey(address))
	if err != nil {
		return nil, err
	}
	ret, err := tcrypto.UnmarshalDKShare(entry.Value, maskPrivate)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func ExistDKShareInRegistry(addr *address.Address) (bool, error) {
	dbase, err := database.GetDB()
	if err != nil {
		return false, err
	}
	return dbase.Contains(dbKey(addr))
}
