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

func (msg *PrivateMessage) OnFirstEmission(state *State) {
	state.addPrivateMessage(msg)
}

func (msg *PrivateMessage) OnReception(state *State, sendReply func(*GossipPacket)) {
	fmt.Println("PRIVATE", msg)
	state.addPrivateMessage(msg)
}

type PointToPoint interface {
	GetOrigin() string
	GetDestination() string
	NextHop() bool
	ToPacket() *GossipPacket
	OnFirstEmission(*State)
	OnReception(*State, func(*GossipPacket))
}

type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}
type GossipPacket struct {
	Simple      *SimpleMessage
	Rumor       *RumorMessage
	Status      *StatusPacket
	Private     *PrivateMessage
	DataRequest *DataRequest
	DataReply   *DataReply
}

func NewDataRequest(origin string, destination string, hash []byte) *DataRequest {
	o := DataRequest{
		Origin:      origin,
		Destination: destination,
		HashValue:   hash,
		HopLimit:    10,
	}
	return &o
}

func (msg *DataRequest) ToPacket() *GossipPacket {
	return &GossipPacket{DataRequest: msg}
}

func (msg *DataRequest) GetOrigin() string {
	return msg.Origin
}

func (msg *DataRequest) GetDestination() string {
	return msg.Destination
}

func (msg *DataRequest) NextHop() bool {
	msg.HopLimit -= 1
	if msg.HopLimit <= 0 {
		return false
	} else {
		return true
	}
}

func (msg *DataRequest) OnFirstEmission(state *State) {
}

func (msg *DataRequest) OnReception(state *State, sendReply func(*GossipPacket)) {
	if _, data := ReadFileForHash(msg.HashValue); len(data) > 0 {
		reply := NewDataReply(msg.Origin, msg.Destination, msg.HashValue, data)
		sendReply(reply.ToPacket())
	}
}

func (msg *DataReply) ToPacket() *GossipPacket {
	return &GossipPacket{DataReply: msg}
}

func (msg *DataReply) GetOrigin() string {
	return msg.Origin
}

func (msg *DataReply) GetDestination() string {
	return msg.Destination
}

func NewDataReply(origin string, destination string, hash []byte, data []byte) *DataReply {
	o := DataReply{
		Origin:      origin,
		Destination: destination,
		HashValue:   hash,
		HopLimit:    10,
		Data:        data,
	}
	return &o
}

func (msg *DataReply) NextHop() bool {
	msg.HopLimit -= 1
	if msg.HopLimit <= 0 {
		return false
	} else {
		return true
	}
}

func (msg *DataReply) OnFirstEmission(state *State) {
}

func (msg *DataReply) OnReception(state *State, sendReply func(*GossipPacket)) {
	state.DispatchDataAck(msg.Origin, HashToUid(msg.HashValue), *msg)
}
