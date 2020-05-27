package registry

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/database"
	"go.dedis.ch/kyber/v3"
)

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
	dbase := database.GetPartition(ks.Address)
	dbkey := database.MakeKey(database.ObjectTypeDistributedKeyData)
	exists, err := dbase.Has(dbkey)
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
	return dbase.Set(dbkey, buf.Bytes())
}

func LoadDKShare(addr *address.Address, maskPrivate bool) (*tcrypto.DKShare, error) {
	data, err := database.GetPartition(addr).Get(database.MakeKey(database.ObjectTypeDistributedKeyData))
	if err != nil {
		return nil, err
	}
	ret, err := tcrypto.UnmarshalDKShare(data, maskPrivate)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func ExistDKShareInRegistry(addr *address.Address) (bool, error) {
	return database.GetPartition(addr).Has(database.MakeKey(database.ObjectTypeDistributedKeyData))
}
