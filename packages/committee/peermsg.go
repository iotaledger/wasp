package committee

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

func (msg *NotifyReqMsg) Write(w io.Writer) error {
	if err := util.WriteUint32(w, msg.StateIndex); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(msg.RequestIds))); err != nil {
		return err
	}
	for _, reqid := range msg.RequestIds {
		if _, err := w.Write(reqid.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (msg *NotifyReqMsg) Read(r io.Reader) error {
	err := util.ReadUint32(r, &msg.StateIndex)
	if err != nil {
		return err
	}
	var arrLen uint16
	err = util.ReadUint16(r, &arrLen)
	if err != nil {
		return err
	}
	if arrLen == 0 {
		return nil
	}
	msg.RequestIds = make([]*sctransaction.RequestId, arrLen)
	for i := range msg.RequestIds {
		msg.RequestIds[i] = new(sctransaction.RequestId)
		_, err = r.Read(msg.RequestIds[i].Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (msg *StartProcessingReqMsg) Write(w io.Writer) error {
	if err := util.WriteUint32(w, msg.StateIndex); err != nil {
		return err
	}
	if _, err := w.Write(msg.RequestId.Bytes()); err != nil {
		return err
	}
	if _, err := w.Write(msg.RewardAddress.Bytes()); err != nil {
		return err
	}
	if err := waspconn.WriteBalances(w, msg.Balances); err != nil {
		return err
	}
	return nil
}

func (msg *StartProcessingReqMsg) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &msg.StateIndex); err != nil {
		return err
	}
	msg.RequestId = new(sctransaction.RequestId)
	if _, err := r.Read(msg.RequestId.Bytes()); err != nil {
		return err
	}
	msg.RewardAddress = new(address.Address)
	if _, err := r.Read(msg.RewardAddress.Bytes()); err != nil {
		return err
	}
	var err error
	if msg.Balances, err = waspconn.ReadBalances(r); err != nil {
		return err
	}
	return nil
}

func (msg *SignedHashMsg) Write(w io.Writer) error {
	if err := util.WriteUint32(w, msg.StateIndex); err != nil {
		return err
	}
	if err := util.WriteTime(w, msg.OrigTimestamp); err != nil {
		return err
	}
	if _, err := w.Write(msg.RequestId.Bytes()); err != nil {
		return err
	}
	if _, err := w.Write(msg.EssenceHash.Bytes()); err != nil {
		return err
	}
	if err := util.WriteBytes16(w, msg.SigShare); err != nil {
		return err
	}
	return nil
}

func (msg *SignedHashMsg) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &msg.StateIndex); err != nil {
		return err
	}
	if err := util.ReadTime(r, &msg.OrigTimestamp); err != nil {
		return err
	}
	if _, err := r.Read(msg.RequestId.Bytes()); err != nil {
		return err
	}
	if _, err := r.Read(msg.EssenceHash.Bytes()); err != nil {
		return err
	}
	var err error
	if msg.SigShare, err = util.ReadBytes16(r); err != nil {
		return err
	}
	return nil
}

func (msg *GetStateUpdateMsg) Write(w io.Writer) error {
	return util.WriteUint32(w, msg.StateIndex)
}

func (msg *GetStateUpdateMsg) Read(r io.Reader) error {
	return util.ReadUint32(r, &msg.StateIndex)
}

func (msg *StateUpdateMsg) Write(w io.Writer) error {
	if err := util.WriteUint32(w, msg.StateIndex); err != nil {
		return err
	}
	if err := msg.StateUpdate.Write(w); err != nil {
		return err
	}
	return util.WriteBoolByte(w, msg.FromVM)
}

func (msg *StateUpdateMsg) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &msg.StateIndex); err != nil {
		return err
	}
	msg.StateUpdate = state.NewStateUpdate(nil, 0)
	if err := msg.StateUpdate.Read(r); err != nil {
		return err
	}
	return util.ReadBoolByte(r, &msg.FromVM)
}

func (msg *TestTraceMsg) Write(w io.Writer) error {
	if !util.ValidPermutation(msg.Sequence) {
		panic(fmt.Sprintf("Write: wrong permutation %+v", msg.Sequence))
	}
	if err := util.WriteUint64(w, uint64(msg.InitTime)); err != nil {
		return err
	}
	if err := util.WriteUint16(w, msg.InitPeerIndex); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(msg.Sequence))); err != nil {
		return err
	}
	for _, idx := range msg.Sequence {
		if err := util.WriteUint16(w, idx); err != nil {
			return err
		}
	}
	if err := util.WriteUint16(w, msg.NumHops); err != nil {
		return err
	}
	return nil
}

func (msg *TestTraceMsg) Read(r io.Reader) error {
	var initTime uint64
	if err := util.ReadUint64(r, &initTime); err != nil {
		return err
	}
	msg.InitTime = int64(initTime)
	if err := util.ReadUint16(r, &msg.InitPeerIndex); err != nil {
		return err
	}
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	msg.Sequence = make([]uint16, size)
	for i := range msg.Sequence {
		if err := util.ReadUint16(r, &msg.Sequence[i]); err != nil {
			return err
		}
	}
	if err := util.ReadUint16(r, &msg.NumHops); err != nil {
		return err
	}
	if !util.ValidPermutation(msg.Sequence) {
		panic(fmt.Sprintf("Read: wrong permutation %+v", msg.Sequence))
	}
	return nil
}
