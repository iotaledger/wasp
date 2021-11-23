package iscp

type RequestTarget struct {
	Contract   Hname
	EntryPoint Hname
}

func NewRequestTarget(contract, entryPoint Hname) RequestTarget {
	return RequestTarget{
		Contract:   contract,
		EntryPoint: entryPoint,
	}
}

func (t RequestTarget) Equals(otherTarget RequestTarget) bool {
	return t.Contract == otherTarget.Contract && t.EntryPoint == otherTarget.EntryPoint
}
