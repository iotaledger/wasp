package peering

type Metrics interface {
	PeerCount(peerCount int)
	RecvEnqueued(messageSize, newPipeSize int)
	RecvDequeued(messageSize, newPipeSize int)
	SendEnqueued(messageSize, newPipeSize int)
	SendDequeued(messageSize, newPipeSize int)
}

type emptyMetrics struct{}

func NewEmptyMetrics() Metrics                                  { return &emptyMetrics{} }
func (*emptyMetrics) PeerCount(peerCount int)                   {}
func (*emptyMetrics) RecvEnqueued(messageSize, newPipeSize int) {}
func (*emptyMetrics) RecvDequeued(messageSize, newPipeSize int) {}
func (*emptyMetrics) SendEnqueued(messageSize, newPipeSize int) {}
func (*emptyMetrics) SendDequeued(messageSize, newPipeSize int) {}
