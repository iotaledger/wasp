package dto

type PeeringNodeStatus struct {
	IsAlive  bool
	NetID    string
	NumUsers int
	PubKey   string
}
