package lib

import (
	"fmt"
	"github.com/dedis/protobuf"
	"net"
)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type GossipPacket struct {
	Simple *SimpleMessage
}

type Gossiper struct {
	address *net.UDPAddr
	conn    *net.UDPConn
	Name    string
}

type Peer struct {
	Address *net.UDPAddr
	Conn    *net.UDPConn
}

func SendGossip(conn *net.UDPConn, msg *GossipPacket) (int, error) {
	packetBytes, err := protobuf.Encode(msg)
	if err != nil {
		return -1, err
	}
	return conn.Write(packetBytes)
}

func SendGossipTo(conn *net.UDPConn, msg *GossipPacket, address *net.UDPAddr) (int, error) {
	packetBytes, err := protobuf.Encode(msg)
	if err != nil {
		return -1, err
	}
	return conn.WriteToUDP(packetBytes, address)
}

func ReceiveGossip(conn *net.UDPConn) (*GossipPacket, error) {
	buffer := make([]byte, 65536)
	bytes_read, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}
	data := buffer[:bytes_read]
	packet := &GossipPacket{}
	err = protobuf.Decode(data, packet)
	return packet, err
}

func AddrOfString(address string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", address)
}

/* StringOfAddr make a string of an net.UDPAddr address.
Warning: this is not the inverse of StringToAddr */
func StringOfAddr(addr *net.UDPAddr) string {
	return addr.IP.String() + ":" + fmt.Sprintf("%v", addr.Port)
}

func OpenPermanentConnection(address string) (*net.UDPConn, *net.UDPAddr, error) {
	udpAddr, err := AddrOfString(address)
	if err != nil {
		return nil, nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	return udpConn, udpAddr, err
}

func OpenWriteConnection(address string) (*net.UDPConn, *net.UDPAddr, error) {
	udpAddr, err := AddrOfString(address)
	if err != nil {
		return nil, nil, err
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	return udpConn, udpAddr, err
}

func NewGossiper(address, name string) (*Gossiper, error) {
	udpConn, udpAddr, err := OpenPermanentConnection(address)
	return &Gossiper{
		address: udpAddr,
		conn:    udpConn,
		Name:    name,
	}, err
}
