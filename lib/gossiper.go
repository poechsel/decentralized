package lib

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"net"
	"sync/atomic"
	"time"
)

var Send_queue = make(NetChannel)

type Gossiper struct {
	Address *net.UDPAddr
	Name    string
	Conn    *net.UDPConn

	/* use an atomic to increment it and get the value */
	CurrentMsgId *uint32

	SimpleMode bool
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

func (gossip *Gossiper) SendRumor(rumor *RumorMessage, address *net.UDPAddr, c NetChannel) {
	fmt.Println("MONGERING with", address)
	gossip.SendPacket(&GossipPacket{Rumor: rumor}, address, c)
}
func (gossip *Gossiper) SendStatus(status *StatusPacket, address *net.UDPAddr, c NetChannel) {
	gossip.SendPacket(&GossipPacket{Status: status}, address, c)
}

func (gossip *Gossiper) SendPacket(msg *GossipPacket, address *net.UDPAddr, c NetChannel) {
	c <- Packet{Address: address, Content: msg}
	//return gossip.Conn.WriteToUDP(packetBytes, address)
}

func NewGossiper(address, name string, simple bool) (*Gossiper, error) {
	udpConn, udpAddr, err := OpenPermanentConnection(address)
	id := uint32(0)
	return &Gossiper{
		Address:      udpAddr,
		Conn:         udpConn,
		Name:         name,
		CurrentMsgId: &id,
		SimpleMode:   simple,
	}, err
}

func (server *Gossiper) ClientHandler(state *State, request Packet) {
	packet := request.Content
	if packet.Simple != nil {
		fmt.Println("CLIENT MESSAGE", packet.Simple.Contents)
		fmt.Println("PEERS", state)
		if server.SimpleMode {
			go server.Broadcast(
				"",
				state,
				&SimpleMessage{
					OriginalName:  server.Name,
					RelayPeerAddr: server.Address.String(),
					Contents:      packet.Simple.Contents})
		} else {
			r := RumorMessage{
				Origin: server.Name,
				ID:     server.NewMsgId(),
				Text:   packet.Simple.Contents}
			go server.HandleRumor(state, server.Address.String(), &r)
		}
	}
}

func (server *Gossiper) ServerHandler(state *State, request Packet) {
	packet := request.Content
	source_string := request.Address.String()
	if source_string != server.Address.String() {
		go state.AddPeer(source_string)
	}
	if packet.Simple != nil {
		fmt.Println("SIMPLE MESSAGE", packet.Simple)
		server.Broadcast(
			request.Address.String(),
			state,
			&SimpleMessage{
				OriginalName:  packet.Simple.OriginalName,
				RelayPeerAddr: server.Address.String(),
				Contents:      packet.Simple.Contents})
	} else if packet.Status != nil {
		fmt.Println("STATUS from", source_string, packet.Status)
		// a status message can either be dispatched and use as an ack
		// or in the negative be used directly here
		if !state.dispatchStatusToPeer(source_string, packet.Status) {
			server.HandleStatus(state, source_string, packet.Status.Want)
		}
	} else if packet.Rumor != nil {
		fmt.Println("RUMOR origin",
			packet.Rumor.Origin, "from",
			source_string, "ID",
			packet.Rumor.ID, "contents",
			packet.Rumor.Text)
		server.HandleRumor(state, source_string, packet.Rumor)
	}
	fmt.Println("PEERS", state)
}

func continue_rumormongering(state *State, address string, server *Gossiper, rumor *RumorMessage) {
	decision := rand.Int() % 2
	//decision := 1
	if decision == 1 {
		random_addr, _, err := state.getRandomPeer(address)
		if err != nil {
			return
		}
		fmt.Println("FLIPPED COIN sending rumor to", random_addr)
		addr, _ := AddrOfString(random_addr)
		server.SendRumor(rumor, addr, Send_queue)
	} else {
		// stop mongering
		return
	}
}

func (server *Gossiper) Broadcast(avoid string, state *State, message *SimpleMessage) {
	state.IterPeers(avoid,
		func(peer *Peer) {
			server.SendPacket(
				&GossipPacket{Simple: message},
				peer.Address,
				Send_queue)
		})
}

func (server *Gossiper) HandleStatus(state *State, address string, remote_status []PeerStatus) bool {
	addr, _ := AddrOfString(address)
	self_status := state.db.GetPeerStatus()
	order, diff_status := CompareStatusVector(self_status, remote_status)
	if order == Status_Self_Knows_More {
		content := state.db.GetMessageContent(diff_status.Identifier, diff_status.NextID)
		rumor := &RumorMessage{
			Origin: diff_status.Identifier,
			ID:     diff_status.NextID,
			Text:   content}
		server.SendRumor(rumor, addr, Send_queue)
		return true
	} else if order == Status_Remote_Knows_More {
		server.SendStatus(&StatusPacket{Want: self_status}, addr, Send_queue)
		return true
	} else {
		fmt.Println("IN SYNC WITH", address)
		return false
	}
}

func (server *Gossiper) HandleRumor(state *State, sender_addr_string string, rumor *RumorMessage) {

	message_added := state.addRumorMessage(rumor, sender_addr_string)

	// send the ack
	if sender_addr_string != server.Address.String() {
		sender_addr, _ := AddrOfString(sender_addr_string)
		self_status := state.db.GetPeerStatus()
		server.SendStatus(&StatusPacket{Want: self_status}, sender_addr, Send_queue)
	}

	if message_added {
		rand_peer_address, rand_peer, err := state.getRandomPeer(sender_addr_string)
		if err != nil {
			continue_rumormongering(state, sender_addr_string, server, rumor)
		} else {
			addr, _ := AddrOfString(rand_peer_address)
			server.SendRumor(rumor, addr, Send_queue)
			rand_peer.RequestStatus()
			timer := time.NewTicker(time.Second)

			select {
			case <-timer.C:
				rand_peer.CancelRequestStatus()
				timer.Stop()
				continue_rumormongering(state, sender_addr_string, server, rumor)
			case ack := <-rand_peer.Status_channel:
				if !server.HandleStatus(state, rand_peer_address, ack.Want) {
					continue_rumormongering(state, sender_addr_string, server, rumor)
				}
			}
		}
	}
}

func (server *Gossiper) AntiEntropy(state *State) {
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				rand_peer_address, _, err := state.getRandomPeer()
				if err == nil {
					addr, _ := AddrOfString(rand_peer_address)
					self_status := state.db.GetPeerStatus()
					server.SendStatus(
						&StatusPacket{Want: self_status},
						addr,
						Send_queue)
				}
			}
		}
	}()
}
