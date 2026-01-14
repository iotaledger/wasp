package acss

// type OutMessages *OutMessagesV
// type OutMessagesV struct {
// 	Vote             []gpa.PayloadOut[MsgVote]
// 	RBCCEPayload     []gpa.PayloadOut[MsgRBCCEPayload]
// 	ImplicateRecover []gpa.PayloadOut[MsgImplicateRecover]
// 	Bracha           []gpa.PayloadOut[rbc.MsgBracha]
// }

// func (m *OutMessagesV) AddAll(msgs OutMessages) OutMessages {
// 	if msgs != nil {
// 		m.Vote = append(m.Vote, msgs.Vote...)
// 		m.RBCCEPayload = append(m.RBCCEPayload, msgs.RBCCEPayload...)
// 		m.ImplicateRecover = append(m.ImplicateRecover, msgs.ImplicateRecover...)
// 		m.Bracha = append(m.Bracha, msgs.Bracha...)
// 	}
// 	return m
// }

// func NoMessages() *OutMessagesV {
// 	return &OutMessagesV{}
// }

// func ConcatMsgs(msgs ...OutMessages) OutMessages {
// 	res := NoMessages()
// 	for _, m := range msgs {
// 		res.AddAll(m)
// 	}
// 	return res
// }
