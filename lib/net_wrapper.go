package lib

import (
	"github.com/dedis/protobuf"
	"net"
)

type Packet struct {
	Address *net.UDPAddr
	Content *GossipPacket
}

type NetChannel chan Packet

func SendPacket(Conn *net.UDPConn, p Packet) error {
	packet, err := protobuf.Encode(p.Content)
	if err != nil {
		return err
	}
	Conn.WriteToUDP(packet, p.Address)
	return nil
}
