package pipe

func identityFunInt(index int) int {
	return index
}

func identityFunInterface(elem interface{}) interface{} {
	return elem
}

func alwaysTrueFun(index int) bool {
	return true
}
