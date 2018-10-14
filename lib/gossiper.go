package lib

import (
	"errors"
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
	fmt.Println("new msg id", gossip.CurrentMsgId)
	return atomic.AddUint32(gossip.CurrentMsgId, 1)
}

func (gossip *Gossiper) Receive(c NetChannel) error {
	buffer := make([]byte, 65536)
	bytes_read, address, err := gossip.Conn.ReadFromUDP(buffer)
	fmt.Println(address)

	if err != nil {
		return err
	}
	data := buffer[:bytes_read]
	packet := &GossipPacket{}
	err = protobuf.Decode(data, packet)
	c <- Packet{Address: address, Content: packet}
	return err
}

func (gossip *Gossiper) ReceiveLoop(c NetChannel) {
	for {
		gossip.Receive(c)
	}
}

func (gossip *Gossiper) SendPacket(msg *GossipPacket, address *net.UDPAddr, c NetChannel) {
	if msg.Rumor != nil {
		fmt.Println("Sending rumor to ", address)
		if msg.Rumor.Text == "" {
			_ = (errors.New("ztre"))

		}
	}
	if msg.Status != nil {
		fmt.Println("Sending status to ", address)
	}
	c <- Packet{Address: address, Content: msg}
	//return gossip.Conn.WriteToUDP(packetBytes, address)
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
