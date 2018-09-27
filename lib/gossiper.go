package lib

import (
	"fmt"
	"github.com/dedis/protobuf"
	"net"
	"sync/atomic"
)

type Gossiper struct {
	Address       *net.UDPAddr
	Name          string
	Conn          *net.UDPConn
	StringAddress string

	/* use an atomic to increment it and get the value */
	CurrentMsgId *uint32
}

/* return elements starting at 1 as it returns the new value */
func (gossip *Gossiper) NewMsgId() uint32 {
	return atomic.AddUint32(gossip.CurrentMsgId, 1)
}

func (gossip *Gossiper) ReceiveGossip() (*GossipPacket, *net.UDPAddr, error) {
	buffer := make([]byte, 65536)
	bytes_read, address, err := gossip.Conn.ReadFromUDP(buffer)
	fmt.Println(address)

	if err != nil {
		return nil, nil, err
	}
	data := buffer[:bytes_read]
	packet := &GossipPacket{}
	err = protobuf.Decode(data, packet)
	return packet, address, err
}

func (gossip *Gossiper) SendGossipTo(msg *GossipPacket, address *net.UDPAddr) (int, error) {
	packetBytes, err := protobuf.Encode(msg)
	if err != nil {
		return -1, err
	}
	return gossip.Conn.WriteToUDP(packetBytes, address)
}

func NewGossiper(address, name string) (*Gossiper, error) {
	udpConn, udpAddr, err := OpenPermanentConnection(address)
	id := uint32(0)
	return &Gossiper{
		StringAddress: StringOfAddr(udpAddr),
		Address:       udpAddr,
		Conn:          udpConn,
		Name:          name,
		CurrentMsgId:  &id,
	}, err
}
