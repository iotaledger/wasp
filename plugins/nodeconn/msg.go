package nodeconn

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

const (
	// wasp -> node
	WaspSendTransactionCode    = byte(1)
	WaspSendSubscribeCode      = byte(2)
	WaspSendGetTransactionCode = byte(3)
	WaspSendGetBalancesCode    = byte(4)

	// node -> wasp
	WaspRecvTransactionCode = byte(5)
	WaspRecvBalancesCode    = byte(6)
)

type WaspSendTransactionMsg struct {
	Tx *valuetransaction.Transaction
}

type WaspSendSubscribeMsg struct {
	Addresses   []address.Address
	PullBacklog bool
}

type WaspSendGetTransactionMsg struct {
	TxId *valuetransaction.ID
}

type WaspSendGetBalancesMsg struct {
	Address *address.Address
}

type WaspRecvTransactionMsg struct {
	Tx *valuetransaction.Transaction
}

type WaspRecvBalancesMsg struct {
	Address  *address.Address
	Balances map[valuetransaction.ID][]*balance.Balance
}

func (msg *WaspSendTransactionMsg) Encode() []byte {
	return msg.Tx.Bytes()
}

func (msg *WaspSendTransactionMsg) Decode(data []byte) error {
	var err error
	msg.Tx, _, err = valuetransaction.FromBytes(data)
	return err
}

func (msg *WaspSendSubscribeMsg) Encode() []byte {
	var buf bytes.Buffer
	_ = util.WriteUint16(&buf, uint16(len(msg.Addresses)))
	for _, col := range msg.Addresses {
		_, _ = buf.Write(col.Bytes())
	}
	_ = util.WriteBoolByte(&buf, msg.PullBacklog)
	return buf.Bytes()
}

func (msg *WaspSendSubscribeMsg) Decode(data []byte) error {
	rdr := bytes.NewReader(data)
	var size uint16
	if err := util.ReadUint16(rdr, &size); err != nil {
		return err
	}
	msg.Addresses = make([]address.Address, size)
	for i := range msg.Addresses {
		n, err := rdr.Read(msg.Addresses[i].Bytes())
		if err != nil {
			return err
		}
		if n != balance.ColorLength {
			return fmt.Errorf("error while reading 'subscribe' message")
		}
	}
	if err := util.ReadBoolByte(rdr, &msg.PullBacklog); err != nil {
		return err
	}
	return nil
}

func (msg *WaspSendGetTransactionMsg) Encode() []byte {
	return msg.TxId.Bytes()
}

func (msg *WaspSendGetTransactionMsg) Decode(data []byte) error {
	msg.TxId = new(valuetransaction.ID)
	n, err := bytes.NewReader(data).Read(msg.TxId.Bytes())
	if err != nil {
		return err
	}
	if n != valuetransaction.IDLength {
		return fmt.Errorf("error while reading 'get transaction' message")
	}
	return nil
}

func (msg *WaspSendGetBalancesMsg) Encode() []byte {
	return msg.Address.Bytes()
}

func (msg *WaspSendGetBalancesMsg) Decode(data []byte) error {
	a, _, err := address.FromBytes(data)
	if err != nil {
		return err
	}
	msg.Address = &a
	return nil
}

func (msg *WaspRecvTransactionMsg) Encode() []byte {
	return msg.Tx.Bytes()
}

func (msg *WaspRecvTransactionMsg) Decode(data []byte) error {
	var err error
	msg.Tx, _, err = valuetransaction.FromBytes(data)
	return err
}

func (msg *WaspRecvBalancesMsg) Encode() []byte {
	var buf bytes.Buffer
	_, _ = buf.Write(msg.Address.Bytes())
	_ = WriteBalances(&buf, msg.Balances)
	return buf.Bytes()
}

func (msg *WaspRecvBalancesMsg) Decode(data []byte) error {
	rdr := bytes.NewReader(data)
	msg.Address = new(address.Address)
	n, err := rdr.Read(msg.Address.Bytes())
	if err != nil {
		return err
	}
	if n != address.Length {
		return fmt.Errorf("error while decoding 'recv balance' message")
	}

	if msg.Balances, err = ReadBalances(rdr); err != nil {
		return err
	}
	return nil
}

func WriteBalances(w io.Writer, balances map[valuetransaction.ID][]*balance.Balance) error {
	if err := util.WriteUint16(w, uint16(len(balances))); err != nil {
		return err
	}
	for txid, bals := range balances {
		if _, err := w.Write(txid[:]); err != nil {
			return err
		}
		if err := util.WriteUint16(w, uint16(len(bals))); err != nil {
			return err
		}
		for _, b := range bals {
			if _, err := w.Write(b.Color().Bytes()); err != nil {
				return err
			}
			if err := util.WriteUint64(w, uint64(b.Value())); err != nil {
				return err
			}
		}
	}
	return nil
}

func ReadBalances(r io.Reader) (map[valuetransaction.ID][]*balance.Balance, error) {
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return nil, err
	}
	ret := make(map[valuetransaction.ID][]*balance.Balance, size)
	for i := uint16(0); i < size; i++ {
		txid := new(valuetransaction.ID)
		n, err := r.Read(txid.Bytes())
		if err != nil {
			return nil, err
		}
		if n != valuetransaction.IDLength {
			return nil, fmt.Errorf("error while decoding 'recv balance' message")
		}
		var numBals uint16
		if err := util.ReadUint16(r, &numBals); err != nil {
			return nil, err
		}
		lst := make([]*balance.Balance, numBals)
		for i := range lst {
			var color balance.Color
			n, err := r.Read(color[:])
			if err != nil {
				return nil, err
			}
			if n != balance.ColorLength {
				return nil, fmt.Errorf("error while decoding 'recv balance' message")
			}
			var value uint64
			if err := util.ReadUint64(r, &value); err != nil {
				return nil, err
			}
			lst[i] = balance.New(color, int64(value))
		}
		ret[*txid] = lst
	}
	return ret, nil
}
