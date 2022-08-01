package isc

// CallTarget the target representation of the request
type CallTarget struct {
	Contract   Hname
	EntryPoint Hname
}

func NewCallTarget(contract, entryPoint Hname) CallTarget {
	return CallTarget{
		Contract:   contract,
		EntryPoint: entryPoint,
	}
}

func (t CallTarget) Equals(otherTarget CallTarget) bool {
	return t.Contract == otherTarget.Contract && t.EntryPoint == otherTarget.EntryPoint
}
