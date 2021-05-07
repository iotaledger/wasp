package mock_chain

//---------------------------------------------
type MockedPeerGroupProvider struct {
	onLock       func()
	onUnlock     func()
	onNumPeers   func() uint16
	onNumIsAlive func(uint16) bool
	onSendMsg    func(targetPeerIndex uint16, msgType byte, msgData []byte) error
}

func NewMockedPeerGroupProvider() *MockedPeerGroupProvider {
	return &MockedPeerGroupProvider{
		onLock:   func() {},
		onUnlock: func() {},
		onNumPeers: func() uint16 {
			return 0
		},
		onNumIsAlive: func(u uint16) bool {
			return false
		},
		onSendMsg: func(targetPeerIndex uint16, msgType byte, msgData []byte) error {
			return nil
		},
	}
}

func (m *MockedPeerGroupProvider) Lock() {
	m.onLock()
}

func (m *MockedPeerGroupProvider) Unlock() {
	m.onUnlock()
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

func (m *MockedPeerGroupProvider) OnLock(f func()) {
	m.onLock = f
}

func (m *MockedPeerGroupProvider) OnUnlock(f func()) {
	m.onUnlock = f
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
