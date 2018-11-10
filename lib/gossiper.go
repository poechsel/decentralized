package lib

import (
	"fmt"
	"github.com/dedis/protobuf"
	"log"
	"math/rand"
	"net"
	"sync"
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
		p := NewPrivateMessage(
			server.Name,
			packet.Private.Text,
			packet.Private.Destination)
		go server.HandlePointToPointMessage(state, server.Address.String(), &p)
	} else if packet.DataRequest != nil {
		fmt.Println("REQUESTING INDEXING filename", packet.DataRequest.Origin)
		go server.UploadFile(packet.DataRequest.Origin)
	} else if packet.DataReply != nil {
		fmt.Println("REQUESTING filename", packet.DataReply.Origin, "from", packet.DataReply.Destination, "hash", HashToUid(packet.DataReply.HashValue))
		go server.DownloadFile(state,
			packet.DataReply.Destination,
			packet.DataReply.HashValue,
			packet.DataReply.Origin)
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
		server.HandlePointToPointMessage(state, source_string, packet.Private)
	} else if packet.DataReply != nil {
		go server.HandlePointToPointMessage(state, source_string, packet.DataReply)
	} else if packet.DataRequest != nil {
		go server.HandlePointToPointMessage(state, source_string, packet.DataRequest)
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

func (server *Gossiper) SendReplyWaitAnswer(state *State, peer string, hash []byte) DataReply {
	dataRequest := NewDataRequest(server.Name, peer, hash)
	ackr := NewAckRequest()
	state.AddDataAck(peer, HashToUid(hash), *ackr)
	for {
		go server.HandlePointToPointMessage(state, server.Address.String(), dataRequest)
		timeout := time.NewTimer(5 * time.Second)
		select {
		case <-timeout.C:
			continue
		case r := <-ackr.AckChannel:
			ackr.Close()
			return r.(DataReply)
		}
	}
}

//out_file is relative to the download folder
func (server *Gossiper) DownloadFile(state *State, peer string, metahash []byte, out_file string) {
	metafilereply := server.SendReplyWaitAnswer(state, peer, metahash)
	metafile := metafilereply.Data
	fmt.Println("DOWNLOADING metafile of", out_file, "from", peer)
	go WriteMetaFile(metafile)
	nparts := len(metafile) / 32
	var wg sync.WaitGroup
	wg.Add(nparts)

	for i := 0; i < len(metafile); i += 32 {
		go func(i int) {
			hash := metafile[i : i+32]
			chunk := server.SendReplyWaitAnswer(state, peer, hash)
			WriteChunkFile(chunk.Data)
			fmt.Println("DOWNLOADING", out_file, "chunk", i+1, "from", peer)
			wg.Done()
		}(i)
	}
	wg.Wait()
	ReconstructFile(out_file, metafile)
	fmt.Println("RECONSTRUCTED file", out_file)
}

// path is relative to share folder
func (server *Gossiper) UploadFile(path string) {
	metafile := SplitFile(path)
	WriteMetaFile(metafile)
}

func (server *Gossiper) HandlePointToPointMessage(state *State, sender_addr_string string, msg PointToPoint) {
	/* This check is only to make sure that we dispatch private messages
	sent by our current node */
	if msg.GetOrigin() == server.Name {
		go msg.OnFirstEmission(state)
	}

	log.Println("PTP", msg.GetDestination(), "=", server.Name)

	if msg.GetDestination() == server.Name {
		if address_origin, ok := state.getRouteTo(msg.GetOrigin()); ok {
			address_origin_udp, _ := AddrOfString(address_origin)
			go msg.OnReception(
				state,
				func(packet *GossipPacket) {
					log.Println("sending answer back to ", msg.GetOrigin(), address_origin_udp)
					go server.SendPacket(packet, address_origin_udp)
				},
			)
		} else {
			go msg.OnReception(
				state,
				func(packet *GossipPacket) {},
			)
		}
	} else {
		/* we make a shallow copy of msg */
		next_msg := msg
		ok := next_msg.NextHop()
		next_address, ok2 := state.getRouteTo(msg.GetDestination())
		if ok && ok2 {
			address, _ := AddrOfString(next_address)
			server.SendPacket(next_msg.ToPacket(), address)
		}
	}
}

func (server *Gossiper) HandleRumor(state *State, sender_addr_string string, rumor *RumorMessage) {

	message_added, isIdGreater := state.addRumorMessage(rumor, sender_addr_string)

	if isIdGreater {
		state.UpdateRoutingTable(rumor.Origin, sender_addr_string)
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

func (server *Gossiper) createRouteRefresh(state *State) *RumorMessage {
	rm := &RumorMessage{Origin: server.Name, ID: server.NewMsgId(), Text: ""}
	state.db.InsertRumorMessage(rm)
	return rm
}

func (server *Gossiper) RefreshRouteLoop(state *State) {
	if server.Rtimer > 0 {
		rm := server.createRouteRefresh(state)
		server.Broadcast("", state, &GossipPacket{Rumor: rm})
		ticker := time.NewTicker(time.Duration(server.Rtimer) * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					rand_peer_address, _, err := state.getRandomPeer()
					if err == nil {
						addr, _ := AddrOfString(rand_peer_address)
						rm := server.createRouteRefresh(state)
						server.SendRumor(rm, addr)
					}
				}
			}
		}()
	}
}
