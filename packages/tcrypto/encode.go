package tcrypto

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"io"
)

func (ks *DKShare) Write(w io.Writer) error {
	_, err := w.Write(ks.Address.Bytes())
	if err != nil {
		return err
	}
	err = util.WriteUint16(w, ks.N)
	if err != nil {
		return err
	}
	err = util.WriteUint16(w, ks.T)
	if err != nil {
		return err
	}
	err = util.WriteUint16(w, ks.Index)
	if err != nil {
		return err
	}
	err = util.WriteUint16(w, uint16(len(ks.PubKeys)))
	if err != nil {
		return err
	}
	for _, pk := range ks.PubKeys {
		pkdata, err := pk.MarshalBinary()
		if err != nil {
			return err
		}
		err = util.WriteBytes16(w, pkdata)
		if err != nil {
			return err
		}
	}
	pkdata, err := ks.priKey.MarshalBinary()
	if err != nil {
		return err
	}
	err = util.WriteBytes16(w, pkdata)
	if err != nil {
		return err
	}
	return nil
}

func (ks *DKShare) Read(r io.Reader) error {
	*ks = DKShare{Suite: bn256.NewSuite()}

	addr := new(address.Address)
	_, err := r.Read(addr[:])
	if err != nil {
		return err
	}
	if addr.Version() != address.VersionBLS {
		return errors.New("not a BLS address")
	}
	var n, t, index uint16
	err = util.ReadUint16(r, &n)
	if err != nil {
		return err
	}
	err = util.ReadUint16(r, &t)
	if err != nil {
		return err
	}
	err = util.ReadUint16(r, &index)
	if err != nil {
		return err
	}
	var num uint16
	err = util.ReadUint16(r, &num)
	if err != nil {
		return err
	}
	pubKeys := make([]kyber.Point, num)
	for i := range pubKeys {
		data, err := util.ReadBytes16(r)
		if err != nil {
			return err
		}
		pubKeys[i] = ks.Suite.G2().Point()
		err = pubKeys[i].UnmarshalBinary(data)
		if err != nil {
			return err
		}
	}
	data, err := util.ReadBytes16(r)
	if err != nil {
		return err
	}
	priKey := ks.Suite.G2().Scalar()
	err = priKey.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	ks.N = n
	ks.T = t
	ks.Index = index
	ks.Address = addr
	ks.PubKeys = pubKeys
	ks.priKey = priKey
	return nil
}

// UnmarshalDKShare parses DKShare, validates and calculates master public key
func UnmarshalDKShare(data []byte, maskPrivate bool) (*DKShare, error) {
	ret := &DKShare{}

	err := ret.Read(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	ret.Aggregated = true
	ret.Committed = true
	ret.PubPoly, err = RecoverPubPoly(ret.Suite, ret.PubKeys, ret.T, ret.N)
	if err != nil {
		return nil, err
	}
	pubKeyOwn := ret.Suite.G2().Point().Mul(ret.priKey, nil)
	if !pubKeyOwn.Equal(ret.PubKeys[ret.Index]) {
		return nil, errors.New("crosscheck I: inconsistency while calculating public key")
	}
	ret.PubKeyOwn = ret.PubKeys[ret.Index]
	ret.PubKeyMaster = ret.PubPoly.Commit()
	if maskPrivate {
		ret.priKey = nil
	}
	binPK, err := ret.PubKeyMaster.MarshalBinary()
	if err != nil {
		return nil, err
	}
	addrFromBin := address.FromBLSPubKey(binPK)
	if !bytes.Equal(addrFromBin.Bytes(), ret.Address.Bytes()) {
		return nil, errors.New("crosscheck II: !HashData(binPK).Equal(ret.ChainID)")
	}
	return ret, nil
}
