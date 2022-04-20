package bracha

import "github.com/iotaledger/wasp/packages/gpa"

type msgPredicateUpdate struct {
	me        gpa.NodeID
	predicate func([]byte) bool
}

var _ gpa.Message = &msgPredicateUpdate{}

func (m *msgPredicateUpdate) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgPredicateUpdate) MarshalBinary() ([]byte, error) {
	panic("not used")
}
