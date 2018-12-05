package lib

import (
	"fmt"
	"github.com/dedis/protobuf"
	"log"
	"math/rand"
	"net"
	"strings"
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

func (gossip *Gossiper) SendSearchRequest(r *SearchRequest, address *net.UDPAddr) {
	gossip.SendPacket(&GossipPacket{SearchRequest: r}, address)
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
		go server.UploadFile(state, packet.DataRequest.Origin)
	} else if packet.DataReply != nil {
		fmt.Println("REQUESTING filename", packet.DataReply.Origin, "from", packet.DataReply.Destination, "hash", HashToUid(packet.DataReply.HashValue))
		go server.DownloadFile(state,
			packet.DataReply.Destination,
			packet.DataReply.HashValue,
			packet.DataReply.Origin)
	} else if packet.SearchRequest != nil {
		go server.LaunchSearch(state, packet.SearchRequest.Keywords, int(packet.SearchRequest.Budget))
	}
}

func (server *Gossiper) ServerHandler(state *State, request Packet) {
	packet := request.Content
	sourceString := request.Address.String()
	if sourceString != server.Address.String() {
		go state.AddPeer(sourceString)
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
		fmt.Println("STATUS from", sourceString, packet.Status)
		// a status message can either be dispatched and use as an ack
		// or in the negative be used directly here
		if !state.dispatchStatusToPeer(sourceString, packet.Status) {
			server.HandleStatus(state, sourceString, packet.Status.Want)
		}
	} else if packet.Rumor != nil {
		fmt.Println("RUMOR origin",
			packet.Rumor.Origin, "from",
			sourceString, "ID",
			packet.Rumor.ID, "contents",
			packet.Rumor.Text)
		server.HandleRumor(state, sourceString, packet.Rumor)
	} else if packet.Private != nil {
		server.HandlePointToPointMessage(state, sourceString, packet.Private)
	} else if packet.DataReply != nil {
		go server.HandlePointToPointMessage(state, sourceString, packet.DataReply)
	} else if packet.DataRequest != nil {
		go server.HandlePointToPointMessage(state, sourceString, packet.DataRequest)
	} else if packet.SearchRequest != nil {
		go server.HandleSearchRequest(state, sourceString, packet.SearchRequest)
	} else if packet.SearchReply != nil {
		go server.HandlePointToPointMessage(state, sourceString, packet.SearchReply)
	}
	fmt.Println("PEERS", state)
}

func (server *Gossiper) RumorMonger(state *State, address string, rumor *RumorMessage) {
	decision := rand.Int() % 2
	if decision == 1 {
		randPeer, err := state.getRandomPeer(address)
		if err != nil {
			return
		}
		fmt.Println("FLIPPED COIN sending rumor to", randPeer.Address)
		server.SendRumor(rumor, randPeer.Address)
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

// out_file is relative to the download folder
// If peer is "" we will use our fileKnowledgeDb to select a good peer
func (server *Gossiper) DownloadFile(state *State, peer string, metahash []byte, out_file string) {
	peerMetaHash := state.FileKnowledgeDB.SelectPeerForMetaHash(peer, HashToUid(metahash))
	metafilereply := server.SendReplyWaitAnswer(state, peerMetaHash, metahash)
	metafile := metafilereply.Data
	fmt.Println("DOWNLOADING metafile of", out_file, "from", peerMetaHash)
	go WriteMetaFile(metafile)
	metahashstring := GetMetaHash(metafile)
	nparts := len(metafile) / 32
	state.FileManager.AddFile(out_file, metahashstring, uint64(nparts))
	var wg sync.WaitGroup
	wg.Add(nparts)

	for i := 0; i < len(metafile); i += 32 {
		go func(i int) {
			hash := metafile[i : i+32]
			chunkhashstring := HashToUid(hash)
			// here we do a conversion: chunks are counted starting 1
			peerChunk := state.FileKnowledgeDB.SelectPeerForChunk(peer, metahashstring, i/32+1)
			chunk := server.SendReplyWaitAnswer(state, peerChunk, hash)
			WriteChunkFile(chunk.Data)
			state.FileManager.AddChunk(metahashstring, chunkhashstring, uint64(i/32+1))
			fmt.Println("DOWNLOADING", out_file, "chunk", i+1, "from", peerChunk)
			wg.Done()
		}(i)
	}
	wg.Wait()
	ReconstructFile(out_file, metafile)
	fmt.Println("RECONSTRUCTED file", out_file)
}

// path is relative to share folder
func (server *Gossiper) UploadFile(state *State, path string) {
	metafile := SplitFile(path)

	metahashstring := GetMetaHash(metafile)
	log.Println("UPLOADED ", path, metahashstring)
	WriteMetaFile(metafile)
	state.FileManager.AddFile(path, metahashstring, uint64(len(metafile)/64))
	for i := 0; i < len(metafile); i += 32 {
		hash := metafile[i : i+32]
		chunkhashstring := HashToUid(hash)
		state.FileManager.AddChunk(metahashstring, chunkhashstring, uint64(i/32+1))
	}
}

func (server *Gossiper) HandleSearchRequest(state *State, senderAddrString string, msg *SearchRequest) {
	if !state.searchRequestCacher.CanTreat(msg) {
		return
	}
	log.Println(server.Name, "Success in handling seach request from", msg)
	pattern := SearchPatternToRegex(strings.Join(msg.Keywords, ","))

	// get our current search result and send them back to the origin
	searchResultSelf := state.FileManager.toSearchReply(pattern)
	log.Println(server.Name, strings.Join(msg.Keywords, ","), searchResultSelf)
	searchReply := NewSearchReply(server.Name, msg.Origin, searchResultSelf)
	go server.HandlePointToPointMessage(state, server.Address.String(), searchReply)

	// remove one from the budget
	budget := int(msg.Budget) - 1

	// If we still got some budget remaining
	if budget > 0 {
		// this will work as we will get back at most len(peers) peers if budget is too big
		randPeers, err := state.getNRandomPeer(budget, server.Address.String())

		if err == nil {
			baseBudgetPerPeer := budget / len(randPeers)
			remainderBudget := budget % len(randPeers)

			for i, nextPeer := range randPeers {
				newBudget := baseBudgetPerPeer
				if i < remainderBudget {
					newBudget += 1
				}
				nextRequest := NewSearchRequest(msg.Origin, uint64(newBudget), msg.Keywords)
				log.Println(server.Name, "sending next to", *nextPeer)
				go server.SendSearchRequest(nextRequest, nextPeer.Address)
			}
		}
	}
}

func (server *Gossiper) LaunchSearch(state *State, keywords []string, budget int) {
	receiveFileChan := make(chan (*SearchResultFrom), 256)
	kstr := strings.Join(keywords, ",")

	currentBudget := budget
	if currentBudget < 2 {
		currentBudget = 2
	}

	fmt.Println("SEARCHING for keywords", strings.Join(keywords, ","), "with budget", budget)
	log.Println("SEARCHING for keywords", strings.Join(keywords, ","), "with budget", budget)

	ticker := time.NewTicker(time.Second)
	nResults := 0
	results := make(map[SearchAnswer]bool)

	uidSearch := state.searchRequestCacher.OpenSearch(keywords, receiveFileChan)
	searchMerger := NewSearchMerger()

	for currentBudget <= 32 && nResults < 2 {
		select {
		case <-ticker.C:
			searchRequest := NewSearchRequest(server.Name, uint64(currentBudget), keywords)
			log.Println("new search request", kstr, currentBudget)
			go server.HandleSearchRequest(state, server.Address.String(), searchRequest)
			currentBudget *= 2

		case result := <-receiveFileChan:
			fmt.Println(result.String())
			log.Println("MATCH", result.From, result.Result.FileName)
			log.Println(result.String())
			for _, chunkId := range result.Result.ChunkMap {
				state.FileKnowledgeDB.Insert(HashToUid(result.Result.MetafileHash), int(chunkId), result.From)
			}
			if searchMerger.mergeResult(result.From, result.Result) {
				/* We overwrite the previous answer if it exits.
				In that case, we know that we have several matches on this file */
				key := SearchAnswer{FileName: result.Result.FileName, MetaHash: HashToUid(result.Result.MetafileHash)}
				results[key] = true
				nResults += 1
			}
		}
	}
	state.searchRequestCacher.CloseSearch(uidSearch)
	fmt.Println("SEARCH FINISHED")
}

func (server *Gossiper) HandlePointToPointMessage(state *State, senderAddrString string, msg PointToPoint) {
	/* This check is only to make sure that we dispatch private messages
	sent by our current node */
	if msg.GetOrigin() == server.Name {
		go msg.OnFirstEmission(state)
	}

	if msg.GetDestination() == server.Name {
		if address_origin, ok := state.getRouteTo(msg.GetOrigin()); ok {
			address_origin_udp, _ := AddrOfString(address_origin)
			go msg.OnReception(
				state,
				func(packet *GossipPacket) {
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

func (server *Gossiper) HandleRumor(state *State, senderAddrString string, rumor *RumorMessage) {

	message_added, isIdGreater := state.addRumorMessage(rumor, senderAddrString)

	if isIdGreater {
		state.UpdateRoutingTable(rumor.Origin, senderAddrString)
	}

	// send the ack
	if senderAddrString != server.Address.String() {
		sender_addr, _ := AddrOfString(senderAddrString)
		self_status := state.db.GetPeerStatus()
		server.SendStatus(&StatusPacket{Want: self_status}, sender_addr)
	}

	/* If we added a message, we then wait for an ack and
	rumormonger if needed */
	if message_added {
		randPeer, err := state.getRandomPeer(senderAddrString)
		if err != nil {
			server.RumorMonger(state, senderAddrString, rumor)
		} else {
			server.SendRumor(rumor, randPeer.Address)
			randPeer.RequestStatus()
			timer := time.NewTicker(time.Second)

			select {
			case <-timer.C:
				randPeer.CancelRequestStatus()
				timer.Stop()
				server.RumorMonger(state, senderAddrString, rumor)
			case ack := <-randPeer.Status_channel:
				if !server.HandleStatus(state, randPeer.Address.String(), ack.Want) {
					server.RumorMonger(state, senderAddrString, rumor)
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
				randPeer, err := state.getRandomPeer()
				if err == nil {
					self_status := state.db.GetPeerStatus()
					server.SendStatus(
						&StatusPacket{Want: self_status},
						randPeer.Address)
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
					randPeer, err := state.getRandomPeer()
					if err == nil {
						rm := server.createRouteRefresh(state)
						server.SendRumor(rm, randPeer.Address)
					}
				}
			}
		}()
	}
}
