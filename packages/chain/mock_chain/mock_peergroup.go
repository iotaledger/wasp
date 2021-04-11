package mock_chain

//---------------------------------------------
type MockedPeerGroupProvider struct {
	onNumPeers                 func() uint16
	onNumIsAlive               func(uint16) bool
	onSendMsg                  func(targetPeerIndex uint16, msgType byte, msgData []byte) error
	onSendToAllUntilFirstError func(msgType byte, msgData []byte) uint16
}

func NewMockedPeerGroup() *MockedPeerGroupProvider {
	return &MockedPeerGroupProvider{
		onNumPeers: func() uint16 {
			return 0
		},
		onNumIsAlive: func(u uint16) bool {
			return false
		},
		onSendMsg: func(targetPeerIndex uint16, msgType byte, msgData []byte) error {
			return nil
		},
		onSendToAllUntilFirstError: func(msgType byte, msgData []byte) uint16 {
			return 0
		},
	}
}

func (m *MockedPeerGroupProvider) NumPeers() uint16 {
	return m.onNumPeers()
}

func (m *MockedPeerGroupProvider) NumIsAlive(quorum uint16) bool {
	return m.onNumIsAlive(quorum)
}

func (m *MockedPeerGroupProvider) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	return m.onSendMsg(targetPeerIndex, msgType, msgData)
}

func (m *MockedPeerGroupProvider) SendToAllUntilFirstError(msgType byte, msgData []byte) uint16 {
	return m.onSendToAllUntilFirstError(msgType, msgData)
}

func (m *MockedPeerGroupProvider) OnNumPeers(f func() uint16) {
	m.onNumPeers = f
}

func (m *MockedPeerGroupProvider) OnNumIsAlive(f func(uint16) bool) {
	m.onNumIsAlive = f
}

func (m *MockedPeerGroupProvider) OnSendMsg(f func(targetPeerIndex uint16, msgType byte, msgData []byte) error) {
	m.onSendMsg = f
}

func (m *MockedPeerGroupProvider) OnSendToAllUntilFirstError(f func(msgType byte, msgData []byte) uint16) {
	m.onSendToAllUntilFirstError = f
}
