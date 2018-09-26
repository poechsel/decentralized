package lib

import (
	"github.com/dedis/protobuf"
	"net"
)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

func (msg SimpleMessage) String() string {
	return "origin " + msg.OriginalName + " from " + msg.RelayPeerAddr + " contents " + msg.Contents
}

type GossipPacket struct {
	Simple *SimpleMessage
}

type Gossiper struct {
	Address          *net.UDPAddr
	Name             string
	Conn             *net.UDPConn
	CanonicalAddress string
}

/*
func SendGossipTo(conn *net.UDPConn, msg *GossipPacket, address *net.UDPAddr) (int, error) {
	packetBytes, err := protobuf.Encode(msg)
	if err != nil {
		return -1, err
	}
	return conn.WriteToUDP(packetBytes, address)
}
*/

func (gossip *Gossiper) ReceiveGossip() (*GossipPacket, error) {
	buffer := make([]byte, 65536)
	bytes_read, _, err := gossip.Conn.ReadFromUDP(buffer)

	if err != nil {
		return nil, err
	}
	data := buffer[:bytes_read]
	packet := &GossipPacket{}
	err = protobuf.Decode(data, packet)
	return packet, err
}

func NewGossiper(address, name string) (*Gossiper, error) {
	udpConn, udpAddr, err := OpenPermanentConnection(address)
	return &Gossiper{
		CanonicalAddress: StringOfAddr(udpAddr),
		Address:          udpAddr,
		Conn:             udpConn,
		Name:             name,
	}, err
}
