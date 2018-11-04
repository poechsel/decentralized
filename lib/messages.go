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

/*
func (msg PrivateMessage) NextHop() (PrivateMessage, bool) {
	m := NewPrivateMessage(msg.Origin, msg.Text, msg.Destination)
	m.HopLimit = msg.HopLimit - 1
	return m, m.HopLimit > 0
}
*/

func (msg *PrivateMessage) ToPacket() *GossipPacket {
	return &GossipPacket{Private: msg}
}

func (msg *PrivateMessage) GetOrigin() string {
	return msg.Origin
}

func (msg *PrivateMessage) GetDestination() string {
	return msg.Destination
}
func (msg *PrivateMessage) NextHop() bool {
	msg.HopLimit -= 1
	if msg.HopLimit <= 0 {
		return false
	} else {
		return true
	}
}

func (msg *PrivateMessage) OnFirstEmission(state *State, addr string) {
	state.addPrivateMessage(msg)
}
func (msg *PrivateMessage) OnReception(state *State, addr string) {
	fmt.Println("PRIVATE", msg)
	state.addPrivateMessage(msg)
}

type PointToPoint interface {
	GetOrigin() string
	GetDestination() string
	NextHop() bool
	ToPacket() *GossipPacket
	OnFirstEmission(*State, string)
	OnReception(*State, string)
}
