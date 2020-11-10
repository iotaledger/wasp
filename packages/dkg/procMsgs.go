package dkg

import (
	"bytes"
	"encoding"
	"io"

	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	rabin_vss "go.dedis.ch/kyber/v3/share/vss/rabin"
)

const (
	rabinDealMsgType          byte = 1
	rabinResponseMsgType      byte = 2
	rabinJustificationMsgType byte = 3
	rabinSecretCommitsMsgType byte = 4
)

type msgByteCoder interface {
	MsgType() byte
	Write(io.Writer) error
	Read(io.Reader) error
}

//
// This file contains message types, exchanged between the DKG nodes
// via the peering network.
//

//
//	rabin_dkg.Deal
//
type rabinDealMsg struct {
	deal *rabin_dkg.Deal
}

func (m *rabinDealMsg) MsgType() byte {
	return rabinDealMsgType
}
func (m *rabinDealMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteUint32(w, m.deal.Index); err != nil {
		return err
	}
	if err = writeMarshaled(w, m.deal.Deal.DHKey); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, m.deal.Deal.Signature); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, m.deal.Deal.Nonce); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, m.deal.Deal.Cipher); err != nil {
		return err
	}
	return nil
}
func (m *rabinDealMsg) Read(r io.Reader) error {
	var err error
	if err = util.ReadUint32(r, &m.deal.Index); err != nil {
		return err
	}
	if err = readMarshaled(r, m.deal.Deal.DHKey); err != nil {
		return err
	}
	if m.deal.Deal.Signature, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if m.deal.Deal.Nonce, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if m.deal.Deal.Cipher, err = util.ReadBytes16(r); err != nil {
		return err
	}
	return nil
}
func (m *rabinDealMsg) fromBytes(buf []byte, group kyber.Group) error {
	m.deal = &rabin_dkg.Deal{
		Deal: &rabin_vss.EncryptedDeal{
			DHKey: group.Point(),
		},
	}
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

//
//	rabin_dkg.Response
//
type rabinResponseMsg struct {
	responses []*rabin_dkg.Response
}

func (m *rabinResponseMsg) MsgType() byte {
	return rabinResponseMsgType
}
func (m *rabinResponseMsg) Write(w io.Writer) error {
	var err error
	listLen := uint32(len(m.responses))
	if err = util.WriteUint32(w, listLen); err != nil {
		return err
	}
	for _, r := range m.responses {
		if err = util.WriteUint32(w, r.Index); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, r.Response.SessionID); err != nil {
			return err
		}
		if err = util.WriteUint32(w, r.Response.Index); err != nil {
			return err
		}
		if err = util.WriteBoolByte(w, r.Response.Approved); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, r.Response.Signature); err != nil {
			return err
		}
	}
	return nil
}
func (m *rabinResponseMsg) Read(r io.Reader) error {
	var err error
	var listLen uint32
	if err = util.ReadUint32(r, &listLen); err != nil {
		return err
	}
	m.responses = make([]*rabin_dkg.Response, int(listLen))
	for i := range m.responses {
		response := rabin_dkg.Response{
			Response: &rabin_vss.Response{},
		}
		m.responses[i] = &response
		if err = util.ReadUint32(r, &response.Index); err != nil {
			return err
		}
		if response.Response.SessionID, err = util.ReadBytes16(r); err != nil {
			return err
		}
		if err = util.ReadUint32(r, &response.Response.Index); err != nil {
			return err
		}
		if err = util.ReadBoolByte(r, &response.Response.Approved); err != nil {
			return err
		}
		if response.Response.Signature, err = util.ReadBytes16(r); err != nil {
			return err
		}
	}
	return nil
}
func (m *rabinResponseMsg) fromBytes(buf []byte) error {
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

//
//	rabin_dkg.Justification
//
type rabinJustificationMsg struct {
	group          kyber.Group // Just for un-marshaling.
	justifications []*rabin_dkg.Justification
}

func (m *rabinJustificationMsg) MsgType() byte {
	return rabinJustificationMsgType
}
func (m *rabinJustificationMsg) Write(w io.Writer) error {
	var err error
	jLen := uint32(len(m.justifications))
	if err = util.WriteUint32(w, jLen); err != nil {
		return err
	}
	for _, j := range m.justifications {
		if err = util.WriteUint32(w, j.Index); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, j.Justification.SessionID); err != nil {
			return err
		}
		if err = util.WriteUint32(w, j.Justification.Index); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, j.Justification.Deal.SessionID); err != nil {
			return err
		}
		if err = writePriShare(w, j.Justification.Deal.SecShare); err != nil {
			return err
		}
		if err = writePriShare(w, j.Justification.Deal.RndShare); err != nil {
			return err
		}
		if err = util.WriteUint32(w, j.Justification.Deal.T); err != nil {
			return err
		}
		if err = util.WriteUint32(w, uint32(len(j.Justification.Deal.Commitments))); err != nil {
			return err
		}
		for i := range j.Justification.Deal.Commitments {
			if err = writeMarshaled(w, j.Justification.Deal.Commitments[i]); err != nil {
				return err
			}
		}
		if err = util.WriteBytes16(w, j.Justification.Signature); err != nil {
			return err
		}
	}
	return nil
}
func (m *rabinJustificationMsg) Read(r io.Reader) error {
	var err error
	var jLen uint32
	if err = util.ReadUint32(r, &jLen); err != nil {
		return err
	}
	m.justifications = make([]*rabin_dkg.Justification, int(jLen))
	for i := range m.justifications {
		j := rabin_dkg.Justification{
			Justification: &rabin_vss.Justification{},
		}
		m.justifications[i] = &j
		if err = util.ReadUint32(r, &j.Index); err != nil {
			return err
		}
		if j.Justification.SessionID, err = util.ReadBytes16(r); err != nil {
			return err
		}
		if err = util.ReadUint32(r, &j.Justification.Index); err != nil {
			return err
		}
		if j.Justification.Deal.SessionID, err = util.ReadBytes16(r); err != nil {
			return err
		}
		if err = readPriShare(r, &j.Justification.Deal.SecShare); err != nil {
			return err
		}
		if err = readPriShare(r, &j.Justification.Deal.RndShare); err != nil {
			return err
		}
		if err = util.ReadUint32(r, &j.Justification.Deal.T); err != nil {
			return err
		}
		var commitmentCount uint32
		if err = util.ReadUint32(r, &commitmentCount); err != nil {
			return err
		}
		j.Justification.Deal.Commitments = make([]kyber.Point, int(commitmentCount))
		for i := range j.Justification.Deal.Commitments {
			j.Justification.Deal.Commitments[i] = m.group.Point()
			if err = readMarshaled(r, j.Justification.Deal.Commitments[i]); err != nil {
				return err
			}
		}
		if j.Justification.Signature, err = util.ReadBytes16(r); err != nil {
			return err
		}
	}
	return nil
}
func (m *rabinJustificationMsg) fromBytes(buf []byte, group kyber.Group) error {
	m.group = group
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

//
//
//
type rabinSecretCommitsMsg struct {
	group         kyber.Group // Just for un-marshaling.
	secretCommits *rabin_dkg.SecretCommits
}

func (m *rabinSecretCommitsMsg) MsgType() byte {
	return rabinSecretCommitsMsgType
}
func (m *rabinSecretCommitsMsg) Write(w io.Writer) error {
	// var err error
	// type SecretCommits struct {
	// 	// Index of the Dealer in the list of participants
	// 	Index uint32
	// 	// Commitments generated by the Dealer
	// 	Commitments []kyber.Point
	// 	// SessionID generated by the Dealer tied to the Deal
	// 	SessionID []byte
	// 	// Signature from the Dealer
	// 	Signature []byte
	// }

	return nil // TODO
}
func (m *rabinSecretCommitsMsg) Read(r io.Reader) error {
	// var err error
	return nil // TODO
}
func (m *rabinSecretCommitsMsg) fromBytes(buf []byte, group kyber.Group) error {
	m.group = group
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

//
//	This works for kyber.Point, kyber.Scalar.
//
func writeMarshaled(w io.Writer, val encoding.BinaryMarshaler) error {
	var err error
	var bin []byte
	if bin, err = val.MarshalBinary(); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, bin); err != nil {
		return err
	}
	return nil
}
func readMarshaled(r io.Reader, val encoding.BinaryUnmarshaler) error {
	var err error
	var bin []byte
	if bin, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if err = val.UnmarshalBinary(bin); err != nil {
		return err
	}
	return nil
}

//
// type PriShare struct {
// 	I int          // Index of the private share
// 	V kyber.Scalar // Value of the private share
// }
//
func writePriShare(w io.Writer, val *share.PriShare) error {
	var err error
	if err = util.WriteBoolByte(w, val == nil); err != nil {
		return err
	}
	if val == nil {
		return nil
	}
	if err = util.WriteUint32(w, uint32(val.I)); err != nil {
		return err
	}
	if err = writeMarshaled(w, val.V); err != nil {
		return err
	}
	return nil
}
func readPriShare(r io.Reader, val **share.PriShare) error {
	var err error
	var valNil bool
	if err = util.ReadBoolByte(r, &valNil); err != nil {
		return err
	}
	if valNil {
		*val = nil
	}
	var i uint32
	if err = util.ReadUint32(r, &i); err != nil {
		return err
	}
	(*val).I = int(i)
	if err = readMarshaled(r, (*val).V); err != nil {
		return err
	}
	return nil
}
