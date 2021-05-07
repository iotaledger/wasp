// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"fmt"
	"net"
	"strconv"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type peerNetworkConfig struct {
	ownNetID    string
	peeringPort int
	neighbors   []string
}

func NewPeerNetworkConfig(ownNetID string, peeringPort int, neighbors ...string) (*peerNetworkConfig, error) {
	if err := CheckMyNetID(ownNetID, peeringPort); err != nil {
		return nil, xerrors.Errorf("NewPeerNetworkConfig: %w", err)
	}
	if !util.AllDifferentStrings(neighbors) {
		return nil, xerrors.Errorf("NewPeerNetworkConfig: neighbors must all be different")
	}
	neigh := make([]string, 0, len(neighbors))
	// eliminate own id from the list
	for _, n := range neighbors {
		if n != ownNetID {
			neigh = append(neigh, n)
		}
	}
	return &peerNetworkConfig{
		ownNetID:    ownNetID,
		peeringPort: peeringPort,
		neighbors:   neigh,
	}, nil
}

func (p *peerNetworkConfig) OwnNetID() string {
	return p.ownNetID
}

func (p *peerNetworkConfig) PeeringPort() int {
	return p.peeringPort
}

func (p *peerNetworkConfig) DefaultNeighbors() []string {
	return p.neighbors
}

func (p *peerNetworkConfig) Neighbors(chainID *coretypes.ChainID) []string {
	return p.DefaultNeighbors()
}

// CheckMyNetID checks if NetID from the committee list represents current node.
func CheckMyNetID(myNetID string, configPort int) error {
	sHost, sPort, err := net.SplitHostPort(myNetID)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(sPort)
	if err != nil {
		return err
	}
	if port != configPort {
		return fmt.Errorf("wrong own network port in %s", myNetID)
	}
	myIPs, err := myIPs()
	if err != nil {
		return err
	}
	ips, err := net.LookupIP(sHost)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if ip.IsLoopback() {
			return nil
		}
		for _, myIP := range myIPs {
			if ip.String() == myIP {
				return nil
			}
		}
	}
	return fmt.Errorf("NetID %s doesn't represent current node", myNetID)
}

func myIPs() ([]string, error) {
	ret := make([]string, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			if ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ret = append(ret, ip.String())
		}
	}
	return ret, nil
}
