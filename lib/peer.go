package lib

import (
	"fmt"
	"github.com/dedis/protobuf"
	"net"
)

type Peer struct {
	Address *net.UDPAddr
	Conn    *net.UDPConn
}

func (peer *Peer) SendGossip(msg *GossipPacket) (int, error) {
	packetBytes, err := protobuf.Encode(msg)
	if err != nil {
		return -1, err
	}
	return peer.Conn.Write(packetBytes)

}

func NewPeer(address string) (*Peer, error) {
	udpAddr, err := AddrOfString(address)
	if err != nil {
		return nil, err
	}
	fmt.Println("IMPORTANT", udpAddr)
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	return &Peer{
		Address: udpAddr,
		Conn:    udpConn}, err
}
