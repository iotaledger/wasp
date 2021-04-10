package test_statemgr

//---------------------------------------------
type dummyGroupProvider struct{}

func NewDummyPeerGroup() *dummyGroupProvider {
	return &dummyGroupProvider{}
}

func (*dummyGroupProvider) NumPeers() uint16 {
	return 5
}

func (*dummyGroupProvider) NumIsAlive(quorum uint16) bool {
	return true
}

func (*dummyGroupProvider) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	return nil
}

func (pgpThis *dummyGroupProvider) SendToAllUntilFirstError(msgType byte, msgData []byte) uint16 {
	return pgpThis.NumPeers()
}
