package lib

import (
	"fmt"
	"net"
	"os"
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
	if err != nil {
		return nil, nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	// TODO remove this line
	udpConn.SetReadBuffer(1048576)
	return udpConn, udpAddr, err
}

func ExitIfError(err error) {
	if err != nil {
		fmt.Errorf("[Error]: %g", err)
		os.Exit(1)
	}
}
