package dto

type PeeringNodeStatus struct {
	PubKey   string
	NetID    string
	IsAlive  bool
	NumUsers int
}
