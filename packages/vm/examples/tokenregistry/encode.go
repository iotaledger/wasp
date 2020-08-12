package tokenregistry

import (
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

func (tm *tokenMetadata) Read(r io.Reader) error {
	if err := util.ReadInt64(r, &tm.supply); err != nil {
		return err
	}
	if err := util.ReadAddress(r, &tm.mintedBy); err != nil {
		return err
	}
	if err := util.ReadAddress(r, &tm.owner); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &tm.created); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &tm.updated); err != nil {
		return err
	}
	var err error
	if tm.description, err = util.ReadString16(r); err != nil {
		return err
	}
	if tm.userDefined, err = util.ReadBytes16(r); err != nil {
		return err
	}
	return nil
}

func (tm *tokenMetadata) Write(w io.Writer) error {
	if err := util.WriteInt64(w, tm.supply); err != nil {
		return err
	}
	if _, err := w.Write(tm.mintedBy[:]); err != nil {
		return err
	}
	if _, err := w.Write(tm.owner[:]); err != nil {
		return err
	}
	if err := util.WriteInt64(w, tm.created); err != nil {
		return err
	}
	if err := util.WriteInt64(w, tm.updated); err != nil {
		return err
	}
	var err error
	if err = util.WriteString16(w, tm.description); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, tm.userDefined); err != nil {
		return err
	}
	return nil
}
