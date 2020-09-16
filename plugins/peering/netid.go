package peering

import (
	"fmt"
	"net"
	"strconv"

	"github.com/iotaledger/wasp/packages/parameters"
)

// check if network location from the committee list represents current node
func checkMyNetworkID() error {
	shost, sport, err := net.SplitHostPort(MyNetworkId())
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(sport)
	if err != nil {
		return err
	}
	if port != parameters.GetInt(parameters.PeeringPort) {
		return fmt.Errorf("wrong own network port in %s", MyNetworkId())
	}
	myIPs, err := myIPs()
	if err != nil {
		return err
	}
	ips, err := net.LookupIP(shost)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if ip.IsLoopback() {
			return nil
		}
		for _, myIp := range myIPs {
			if ip.String() == myIp {
				return nil
			}
		}
	}
	return fmt.Errorf("network location %s doesn't represent current node", MyNetworkId())
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
