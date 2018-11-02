package lib

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"net"
	"sync/atomic"
	"time"
)

type Gossiper struct {
	Address *net.UDPAddr
	Name    string
	Conn    *net.UDPConn

	/* use an atomic to increment it and get the value */
	CurrentMsgId *uint32

	SimpleMode bool

	Rtimer int
}

/* return elements starting at 1 as it returns the new value */
func (gossip *Gossiper) NewMsgId() uint32 {
	return atomic.AddUint32(gossip.CurrentMsgId, 1)
}

func (gossip *Gossiper) Receive(c NetChannel) error {
	buffer := make([]byte, 65536)
	bytes_read, address, err := gossip.Conn.ReadFromUDP(buffer)

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

func (gossip *Gossiper) SendRumor(rumor *RumorMessage, address *net.UDPAddr) {
	fmt.Println("MONGERING with", address)
	gossip.SendPacket(&GossipPacket{Rumor: rumor}, address)
}
func (gossip *Gossiper) SendPrivate(private *PrivateMessage, address *net.UDPAddr) {
	gossip.SendPacket(&GossipPacket{Private: private}, address)
}
func (gossip *Gossiper) SendStatus(status *StatusPacket, address *net.UDPAddr) {
	gossip.SendPacket(&GossipPacket{Status: status}, address)
}

/* This queue is only present to make sure that Q1 works nearly everytime */
var Send_queue = make(NetChannel)

func (gossip *Gossiper) SendPacket(msg *GossipPacket, address *net.UDPAddr) {
	Send_queue <- Packet{Address: address, Content: msg}
	//	SendPacket(gossip.Conn, Packet{Address: address, Content: msg})
}

func NewGossiper(address, name string, simple bool, rtimer int) (*Gossiper, error) {
	udpConn, udpAddr, err := OpenPermanentConnection(address)
	id := uint32(0)
	return &Gossiper{
		Address:      udpAddr,
		Conn:         udpConn,
		Name:         name,
		CurrentMsgId: &id,
		SimpleMode:   simple,
		Rtimer:       rtimer,
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
				&GossipPacket{Simple: &SimpleMessage{
					OriginalName:  server.Name,
					RelayPeerAddr: server.Address.String(),
					Contents:      packet.Simple.Contents}})
		} else {
			r := RumorMessage{
				Origin: server.Name,
				ID:     server.NewMsgId(),
				Text:   packet.Simple.Contents}
			go server.HandleRumor(state, server.Address.String(), &r)
		}
	} else if packet.Private != nil {
		p := NewPrivateMessage(packet.Private.Origin, packet.Private.Text, packet.Private.Destination)
		go server.HandlePrivateMessage(state, server.Address.String(), &p)
	}
}

func (server *Gossiper) ServerHandler(state *State, request Packet) {
	packet := request.Content
	source_string := request.Address.String()
	if source_string != server.Address.String() {
		go state.AddPeer(source_string)
	}
	fmt.Println(state.GetRoutingTableNames())
	if packet.Simple != nil {
		fmt.Println("SIMPLE MESSAGE", packet.Simple)
		server.Broadcast(
			request.Address.String(),
			state,
			&GossipPacket{Simple: &SimpleMessage{
				OriginalName:  packet.Simple.OriginalName,
				RelayPeerAddr: server.Address.String(),
				Contents:      packet.Simple.Contents}})
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
	} else if packet.Private != nil {
		server.HandlePrivateMessage(state, source_string, packet.Private)
	}
	fmt.Println("PEERS", state)
}

func (server *Gossiper) RumorMonger(state *State, address string, rumor *RumorMessage) {
	decision := rand.Int() % 2
	if decision == 1 {
		random_addr, _, err := state.getRandomPeer(address)
		if err != nil {
			return
		}
		fmt.Println("FLIPPED COIN sending rumor to", random_addr)
		addr, _ := AddrOfString(random_addr)
		server.SendRumor(rumor, addr)
	} else {
		// stop mongering
		return
	}
}

func (server *Gossiper) Broadcast(avoid string, state *State, packet *GossipPacket) {
	state.IterPeers(avoid,
		func(peer *Peer) {
			server.SendPacket(packet, peer.Address)
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
		server.SendRumor(rumor, addr)
		return true
	} else if order == Status_Remote_Knows_More {
		server.SendStatus(&StatusPacket{Want: self_status}, addr)
		return true
	} else {
		fmt.Println("IN SYNC WITH", address)
		return false
	}
}

func (server *Gossiper) HandlePrivateMessage(state *State, sender_addr_string string, private *PrivateMessage) {
	/* This check is only to make sure that we dispatch private messages
	sent by our current node */
	if private.Origin == server.Name {
		state.addPrivateMessage(private)
	}

	if private.Destination == server.Name {
		fmt.Println("PRIVATE", private)
		state.addPrivateMessage(private)
	} else {
		next, ok := private.NextHop()
		next_address, ok2 := state.getRouteTo(private.Destination)
		if ok && ok2 {
			address, _ := AddrOfString(next_address)
			server.SendPrivate(&next, address)
		}
	}
}

func (server *Gossiper) HandleRumor(state *State, sender_addr_string string, rumor *RumorMessage) {

	message_added, isIdGreater := state.addRumorMessage(rumor, sender_addr_string)

	if isIdGreater {
		state.updateRoutingTable(rumor.Origin, sender_addr_string)
	}

	// send the ack
	if sender_addr_string != server.Address.String() {
		sender_addr, _ := AddrOfString(sender_addr_string)
		self_status := state.db.GetPeerStatus()
		server.SendStatus(&StatusPacket{Want: self_status}, sender_addr)
	}

	/* If we added a message, we then wait for an ack and
	rumormonger if needed */
	if message_added {
		rand_peer_address, rand_peer, err := state.getRandomPeer(sender_addr_string)
		if err != nil {
			server.RumorMonger(state, sender_addr_string, rumor)
		} else {
			addr, _ := AddrOfString(rand_peer_address)
			server.SendRumor(rumor, addr)
			rand_peer.RequestStatus()
			timer := time.NewTicker(time.Second)

			select {
			case <-timer.C:
				rand_peer.CancelRequestStatus()
				timer.Stop()
				server.RumorMonger(state, sender_addr_string, rumor)
			case ack := <-rand_peer.Status_channel:
				if !server.HandleStatus(state, rand_peer_address, ack.Want) {
					server.RumorMonger(state, sender_addr_string, rumor)
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
						addr)
				}
			}
		}
	}()
}

func (server *Gossiper) RefreshRouteLoop(state *State) {
	if server.Rtimer > 0 {
		rm := &RumorMessage{Origin: server.Name, ID: server.NewMsgId(), Text: ""}
		server.Broadcast("", state, &GossipPacket{Rumor: rm})
		ticker := time.NewTicker(time.Duration(server.Rtimer) * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					rand_peer_address, _, err := state.getRandomPeer()
					fmt.Println("RANDPEER", rand_peer_address)
					if err == nil {
						addr, _ := AddrOfString(rand_peer_address)
						rm := &RumorMessage{
							Origin: server.Name,
							ID:     server.NewMsgId(),
							Text:   "",
						}
						server.SendRumor(rm, addr)
					}
				}
			}
		}()
	}
}
