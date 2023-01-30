package isc

// CallTarget the target representation of the request
type CallTarget struct {
	Contract   Hname `json:"contract" swagger:"required"`
	EntryPoint Hname `json:"entryPoint" swagger:"required"`
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
