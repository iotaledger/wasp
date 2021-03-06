// +build ignore

package tokenregistry

import (
	"io"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"
)

func (tm *TokenMetadata) Read(r io.Reader) error {
	if err := util.ReadInt64(r, &tm.Supply); err != nil {
		return err
	}
	if err := iscp.ReadAgentID(r, &tm.MintedBy); err != nil {
		return err
	}
	if err := iscp.ReadAgentID(r, &tm.Owner); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &tm.Created); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &tm.Updated); err != nil {
		return err
	}
	var err error
	if tm.Description, err = util.ReadString16(r); err != nil {
		return err
	}
	if tm.UserDefined, err = util.ReadBytes16(r); err != nil {
		return err
	}
	return nil
}

func (tm *TokenMetadata) Write(w io.Writer) error {
	if err := util.WriteInt64(w, tm.Supply); err != nil {
		return err
	}
	if _, err := w.Write(tm.MintedBy[:]); err != nil {
		return err
	}
	if _, err := w.Write(tm.Owner[:]); err != nil {
		return err
	}
	if err := util.WriteInt64(w, tm.Created); err != nil {
		return err
	}
	if err := util.WriteInt64(w, tm.Updated); err != nil {
		return err
	}
	var err error
	if err = util.WriteString16(w, tm.Description); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, tm.UserDefined); err != nil {
		return err
	}
	return nil
}
