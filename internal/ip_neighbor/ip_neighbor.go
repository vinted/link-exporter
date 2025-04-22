package ip_neighbor

import (
	"net"
	"net/netip"

	nl "github.com/vishvananda/netlink"
)

type Neighbor struct {
	InterfaceName string
	MAC           net.HardwareAddr
	IP            netip.Addr
}

func ListNeighbors() []Neighbor {
	var neighbor_list []Neighbor
	nl, _ := nl.NeighList(0, 0)
	ifaces, _ := net.Interfaces()
	for _, neigh := range nl {
		ip, err := netip.ParseAddr(neigh.IP.String())
		if err == nil && ip.Is6() == true {
			if ip.IsLinkLocalMulticast() == false && ip.IsInterfaceLocalMulticast() == false && ip.IsMulticast() == false && ip.IsLoopback() == false && len(neigh.HardwareAddr.String()) > 0 {
				neighbor_list = append(neighbor_list, Neighbor{InterfaceName: indexToInterface(ifaces, neigh.LinkIndex), MAC: neigh.HardwareAddr, IP: ip})
			}
		}
	}
	return neighbor_list
}

func indexToInterface(interfaces []net.Interface, index int) string {
	for _, iface := range interfaces {
		if iface.Index == index {
			return iface.Name
		}
	}
	return ""
}
