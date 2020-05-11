package registry

import "fmt"

type PortAddr struct {
	Port int    `json:"port"`
	Addr string `json:"addr"`
}

func (pa *PortAddr) AdjustedIP() (string, int) {
	if pa.Addr == "localhost" {
		return "127.0.0.1", pa.Port
	}
	return pa.Addr, pa.Port
}

func (pa *PortAddr) String() string {
	return fmt.Sprintf("%s:%d", pa.Addr, pa.Port)
}
