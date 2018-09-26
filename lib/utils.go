package lib

import (
	"fmt"
	"net"
)

func AddrOfString(address string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", address)
}

/* StringOfAddr make a string of an net.UDPAddr address.
Warning: this is not the inverse of StringToAddr */
func StringOfAddr(addr *net.UDPAddr) string {
	return addr.String()
}

func OpenPermanentConnection(address string) (*net.UDPConn, *net.UDPAddr, error) {
	udpAddr, err := AddrOfString(address)
	fmt.Println("Create gossip at ADDRESS: ", udpAddr)
	if err != nil {
		return nil, nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	return udpConn, udpAddr, err
}
