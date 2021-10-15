package pipe

type SimpleHashable int

var _ Hashable = SimpleHashable(0)

func (sh SimpleHashable) GetHash() interface{} {
	return sh
}

func (sh SimpleHashable) Equals(elem interface{}) bool {
	other, ok := elem.(SimpleHashable)
	if !ok {
		return false
	}
	return sh == other
}

//--

func identityFunInt(index int) int {
	return index
}

func alwaysTrueFun(index int) bool {
	return true
}

func priorityFunMod2(i interface{}) bool {
	return priorityFunMod(i, 2)
}

func priorityFunMod3(i interface{}) bool {
	return priorityFunMod(i, 3)
}

func priorityFunMod(i interface{}, mod SimpleHashable) bool {
	return i.(SimpleHashable)%mod == 0
}
