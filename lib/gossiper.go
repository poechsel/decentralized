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

func (msg SimpleMessage) String() string {
	return "origin " + msg.OriginalName + " from " + msg.RelayPeerAddr + " contents " + msg.Contents
}

type GossipPacket struct {
	Simple *SimpleMessage
}

type Gossiper struct {
	Address       *net.UDPAddr
	Name          string
	Conn          *net.UDPConn
	StringAddress string
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

func (gossip *Gossiper) ReceiveGossip() (*GossipPacket, string, error) {
	buffer := make([]byte, 65536)
	bytes_read, address, err := gossip.Conn.ReadFromUDP(buffer)
	fmt.Println(address)

	if err != nil {
		return nil, "", err
	}
	data := buffer[:bytes_read]
	packet := &GossipPacket{}
	err = protobuf.Decode(data, packet)
	return packet, StringOfAddr(address), err
}

func NewGossiper(address, name string) (*Gossiper, error) {
	udpConn, udpAddr, err := OpenPermanentConnection(address)
	return &Gossiper{
		StringAddress: StringOfAddr(udpAddr),
		Address:       udpAddr,
		Conn:          udpConn,
		Name:          name,
	}, err
}
