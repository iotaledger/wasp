package sm_messages

import "github.com/iotaledger/wasp/packages/util/rwutil"

const (
	MsgTypeBlockMessage rwutil.Kind = iota
	MsgTypeGetBlockMessage
)
