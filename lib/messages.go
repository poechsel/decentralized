package lib

import (
	"fmt"
)

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

func (s *StatusPacket) String() string {
	acc := ""
	for _, peer := range s.Want {
		acc += "peer " + peer.Identifier + " nextID " + fmt.Sprint(peer.NextID) + " "
	}
	return acc
}

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

func (msg SimpleMessage) String() string {
	return "origin " + msg.OriginalName + " from " + msg.RelayPeerAddr + " contents " + msg.Contents
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type GossipPacket struct {
	Simple  *SimpleMessage
	Rumor   *RumorMessage
	Status  *StatusPacket
	Private *PrivateMessage
}

func NewPrivateMessage(origin string, text string, destination string) PrivateMessage {
	return PrivateMessage{
		ID:          0,
		Origin:      origin,
		Text:        text,
		Destination: destination,
		HopLimit:    10}
}

func (msg PrivateMessage) String() string {
	return "origin " + msg.Origin + " hop-limit " + fmt.Sprint(msg.HopLimit) + " contents " + msg.Text
}

func (msg PrivateMessage) NextHop() (PrivateMessage, bool) {
	m := NewPrivateMessage(msg.Origin, msg.Text, msg.Destination)
	m.HopLimit = msg.HopLimit - 1
	return msg, m.HopLimit > 0
}
