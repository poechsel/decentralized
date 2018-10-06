package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/poechsel/Peerster/lib"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func ignore(x interface{}) {
	_ = x
}

var send_queue = make(lib.NetChannel)
var client_queue = make(lib.NetChannel)
var msg_queue = make(lib.NetChannel)

type State struct {
	lock_peers  *sync.RWMutex
	known_peers map[string]*lib.Peer
	list_peers  []string
	simple      bool
	db          *lib.Database
}

func (state *State) getRandomPeer(avoid ...string) (string, *lib.Peer, error) {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()

	restricted := make(map[string]bool)
	for _, a := range avoid {
		restricted[a] = true
	}
	var peers []string
	for _, name := range state.list_peers {
		if _, ok := restricted[name]; !ok {
			peers = append(peers, name)
		}
	}

	if len(peers) == 0 {
		return "", nil, errors.New("No peer to select from")
	}
	k := rand.Intn(len(peers))
	name := state.list_peers[k]
	return name, state.known_peers[name], nil
}

func (state *State) dispatchStatusToPeer(status *lib.StatusPacket) bool {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()

	for _, peer := range state.known_peers {
		if peer.DispatchStatus(status) {
			return true
		}
	}
	return false
}

func (state *State) addPeer(address string) bool {
	state.lock_peers.Lock()
	defer state.lock_peers.Unlock()
	if _, ok := state.known_peers[address]; address == "" || ok {
		return false
	} else {
		peer, err := lib.NewPeer(address)
		if err == nil {
			state.known_peers[address] = peer
			state.list_peers = append(state.list_peers, address)
		}
		return true
	}
}

func (state *State) String() string {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()
	fmt.Println("CALLED")
	keys := make([]string, 0, len(state.known_peers))
	for key := range state.known_peers {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}

func (state *State) broadcast(gossiper *lib.Gossiper, message *lib.SimpleMessage, avoid string) {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()
	for addr, peer := range state.known_peers {
		if avoid != addr {
			gossiper.SendPacket(&lib.GossipPacket{Simple: message}, peer.Address, send_queue)
		}
	}
}

func client_handler(state *State, server *lib.Gossiper, request lib.Packet) {
	packet := request.Content
	if packet.Simple != nil {
		fmt.Println("CLIENT MESSAGE", packet.Simple.Contents)
		fmt.Println("PEERS", state)
		if state.simple {
			go state.broadcast(
				server,
				&lib.SimpleMessage{
					OriginalName:  server.Name,
					RelayPeerAddr: server.StringAddress,
					Contents:      packet.Simple.Contents},
				"")
		} else {
			r := lib.RumorMessage{
				Origin: server.Name,
				ID:     server.NewMsgId(),
				Text:   packet.Simple.Contents}
			go handle_rumor(state, server.StringAddress, server, &r)
		}
	}
}

func server_handler(state *State, server *lib.Gossiper, request lib.Packet) {
	packet := request.Content
	source_string := lib.StringOfAddr(request.Address)
	go state.addPeer(source_string)
	if packet.Simple != nil {
		fmt.Println("SIMPLE MESSAGE", packet.Simple)
		fmt.Println("PEERS", state)
		go state.broadcast(
			server,
			&lib.SimpleMessage{
				OriginalName:  packet.Simple.OriginalName,
				RelayPeerAddr: server.StringAddress,
				Contents:      packet.Simple.Contents},
			source_string)
	} else if packet.Status != nil {
		fmt.Printf("STATUS from %v %v\n", source_string, packet.Status)
		// a status message can either be dispatched and use as an ack
		// or in the negative be used directly here
		if !state.dispatchStatusToPeer(packet.Status) {
			fmt.Println("status used normally")
			if !handle_status(state, source_string, server, packet.Status.Want) {

			}
			// anti entropy stuff
		} else {
			fmt.Println("status used as an ack")
		}
	} else if packet.Rumor != nil {
		fmt.Printf("RUMOR origin %v from %v ID %v contents %v\n",
			packet.Rumor.Origin,
			source_string,
			packet.Rumor.ID,
			packet.Rumor.Text)
		if packet.Rumor.Origin != server.Name {
			// oupsi, add the packet to the db
			go func() {
				self_status := state.db.GetPeerStatus()
				message := lib.GossipPacket{Status: &lib.StatusPacket{Want: self_status}}
				go server.SendPacket(&message, request.Address, send_queue)
			}()
			go handle_rumor(state, source_string, server, packet.Rumor)
		}
	}
}

func continue_rumormongering(state *State, address string, server *lib.Gossiper, rumor *lib.RumorMessage) {
	decision := rand.Int() % 2
	if decision == 1 {
		random_addr, _, err := state.getRandomPeer(address)
		if err != nil {
			return
		}
		fmt.Printf("FLIPPED COIN sending rumor to %v\n", random_addr)
		addr, _ := lib.AddrOfString(random_addr)
		fmt.Printf("MONGERING with %v\n", addr)
		go server.SendPacket(&lib.GossipPacket{Rumor: rumor}, addr, send_queue)
	} else {
		// stop mongering
		return
	}
}

func handle_status(state *State, address string, server *lib.Gossiper, remote_status []lib.PeerStatus) bool {
	addr, _ := lib.AddrOfString(address)
	self_status := state.db.GetPeerStatus()
	order, diff_status := lib.CompareStatusVector(self_status, remote_status)
	if order == lib.Status_Self_Knows_More {
		content := state.db.GetMessageContent(diff_status.Identifier, diff_status.NextID)
		fmt.Printf("MONGERING with %v\n", addr)
		message := lib.GossipPacket{Rumor: &lib.RumorMessage{Origin: diff_status.Identifier, ID: diff_status.NextID, Text: content}}
		go server.SendPacket(&message, addr, send_queue)
		return true
	} else if order == lib.Status_Remote_Knows_More {
		message := lib.GossipPacket{Status: &lib.StatusPacket{Want: self_status}}
		go server.SendPacket(&message, addr, send_queue)
		return true
	} else {
		fmt.Printf("IN SYNC with %v\n", address)
	}
	return false
}

func handle_rumor(state *State, address string, server *lib.Gossiper, rumor *lib.RumorMessage) {
	if !state.db.PossessRumorMessage(rumor) {
		go state.db.InsertRumorMessage(rumor)
		rand_peer_address, rand_peer, err := state.getRandomPeer(address)
		if err != nil {
			go continue_rumormongering(state, address, server, rumor)
			return
		}

		addr, _ := lib.AddrOfString(rand_peer_address)
		go server.SendPacket(&lib.GossipPacket{Rumor: rumor}, addr, send_queue)
		rand_peer.RequestStatus()
		timer := time.NewTicker(1 * time.Second)

		go func() {
			select {
			case <-timer.C:
				fmt.Println("TIMEOUT")
				timer.Stop()
				go continue_rumormongering(state, address, server, rumor)
			case ack := <-rand_peer.Status_channel:
				fmt.Println("Got Ack!")
				go func() {
					if !handle_status(state, rand_peer_address, server, ack.Want) {
						go continue_rumormongering(state, address, server, rumor)
					}
				}()
			}
		}()
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	client_port := flag.String("UIPort", "8080", "Port for the UI client")
	gossip_addr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	gossip_name := flag.String("name", "", "name of the gossiper")
	peers_param := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	var simple = flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	flag.Parse()

	peers_list := strings.Split(*peers_param, ",")

	gossiper, err := lib.NewGossiper(*gossip_addr, *gossip_name)
	lib.ExitIfError(err)

	client_server, err := lib.NewGossiper("127.0.0.1:"+*client_port, "client")
	lib.ExitIfError(err)

	db := lib.NewDatabase()
	state := &State{
		known_peers: make(map[string]*lib.Peer),
		db:          &db,
		simple:      *simple, lock_peers: &sync.RWMutex{}}

	for _, peer_addr := range peers_list {
		state.addPeer(peer_addr)
	}

	go gossiper.ReceiveLoop(msg_queue)
	go client_server.ReceiveLoop(client_queue)

	/*
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ticker.C:
					rand_peer_address, _ := state.getRandomPeer()
					addr, _ := lib.AddrOfString(rand_peer_address)
					self_status := state.db.GetPeerStatus()
					message := lib.GossipPacket{Status: &lib.StatusPacket{Want: self_status}}
					go gossiper.SendPacket(&message, addr, send_queue)
				}
			}
		}()
	*/

	/*
		a := lib.NewSparseSequence()
		a.Insert(0)
		a.Insert(1)
		for i := 0; i < 24; i++ {
			a.Insert(uint32(i))
		}
		for i := 24; i < 45; i++ {
			a.Insert(uint32(i))
		}
		a.Print()
		fmt.Println(a.GetMinNotPresent())
	*/
	// infinite loop
	for {
		select {
		case request := <-client_queue:
			client_handler(state, gossiper, request)

		case request := <-msg_queue:
			server_handler(state, gossiper, request)

		case write := <-send_queue:
			// should probably be removed, but ain't nobody got time for that
			lib.SendPacket(gossiper.Conn, write)
		}
	}
}
