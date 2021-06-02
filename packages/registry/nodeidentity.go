// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/key"
)

// GetNodeIdentity implements NodeIdentityProvider.
func (r *Impl) GetNodeIdentity() (*key.Pair, error) {
	var err error
	var pair *key.Pair
	dbKey := dbKeyForNodeIdentity()
	var exists bool
	var data []byte
	exists, err = r.store.Has(dbKey)
	if !exists {
		pair = key.NewKeyPair(r.suite)
		if data, err = keyPairToBytes(pair); err != nil {
			return nil, err
		}
		r.store.Set(dbKey, data)
		r.log.Info("Node identity key pair generated.")
		return pair, nil
	}
	if data, err = r.store.Get(dbKey); err != nil {
		return nil, err
	}
	if pair, err = keyPairFromBytes(data, r.suite); err != nil {
		return nil, err
	}
	return pair, nil
}

// GetNodePublicKey implements NodeIdentityProvider.
func (r *Impl) GetNodePublicKey() (kyber.Point, error) {
	var err error
	var pair *key.Pair
	if pair, err = r.GetNodeIdentity(); err != nil {
		return nil, err
	}
	return pair.Public, nil
}

func dbKeyForNodeIdentity() []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeNodeIdentity)
}

func keyPairToBytes(pair *key.Pair) ([]byte, error) {
	var err error
	var w bytes.Buffer
	if err = util.WriteMarshaled(&w, pair.Private); err != nil {
		return nil, err
	}
	if err = util.WriteMarshaled(&w, pair.Public); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func keyPairFromBytes(buf []byte, suite kyber.Group) (*key.Pair, error) {
	var err error
	r := bytes.NewReader(buf)
	pair := key.Pair{
		Public:  suite.Point(),
		Private: suite.Scalar(),
	}
	if err = util.ReadMarshaled(r, pair.Private); err != nil {
		return nil, err
	}
	if err = util.ReadMarshaled(r, pair.Public); err != nil {
		return nil, err
	}
	return &pair, nil
}
