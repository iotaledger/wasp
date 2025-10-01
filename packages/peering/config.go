// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"time"
)

// CheckPeeringURL verifies if peeringURL is of proper format.
func CheckPeeringURL(url string) error {
	sHost, sPort, err := net.SplitHostPort(url)
	if err != nil {
		return err
	}
	if sHost == "" {
		return errors.New("peeringURL: host part missing")
	}
	port, err := strconv.Atoi(sPort)
	if err != nil {
		return err
	}
	if port == 0 {
		return errors.New("peeringURL: invalid port")
	}
	return nil
}

// CheckMyPeeringURL checks if PeeringURL from the committee list represents current node.
func CheckMyPeeringURL(myPeeringURL string, configPort int) error {
	sHost, sPort, err := net.SplitHostPort(myPeeringURL)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(sPort)
	if err != nil {
		return err
	}
	if port != configPort {
		return fmt.Errorf("wrong own network port in %s", myPeeringURL)
	}
	myIPs, err := myIPs()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, sHost)
	if err != nil {
		return err
	}
	myIPSet := make(map[netip.Addr]bool, len(myIPs))
	for _, s := range myIPs {
		if a, err := netip.ParseAddr(s); err == nil {
			myIPSet[a.Unmap()] = true
		}
	}
	for _, a := range addrs {
		if a.IP.IsLoopback() {
			return nil
		}
		if addr, ok := netip.AddrFromSlice(a.IP); ok {
			if _, hit := myIPSet[addr.Unmap()]; hit {
				return nil
			}
		}
	}
	return fmt.Errorf("peeringURL %s doesn't represent current node", myPeeringURL)
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
