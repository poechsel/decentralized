package main

import (
	"flag"
	"fmt"
	"github.com/poechsel/Peerster/lib"
	"strings"
	"sync"
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
	simple      bool
	db          *lib.Database
}

func (state State) dispatchStatusToPeer(status *lib.StatusPacket) bool {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()

	for _, peer := range state.known_peers {
		if peer.DispatchStatus(status) {
			return true
		}
	}
	return false
}

func (state State) addPeer(address string) bool {
	state.lock_peers.Lock()
	defer state.lock_peers.Unlock()
	if _, ok := state.known_peers[address]; address == "" || ok {
		return false
	} else {
		peer, err := lib.NewPeer(address)
		if err == nil {
			state.known_peers[address] = peer
		}
		return true
	}
}

func (state State) String() string {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()
	fmt.Println("CALLED")
	keys := make([]string, 0, len(state.known_peers))
	for key := range state.known_peers {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}

func (state State) broadcast(gossiper *lib.Gossiper, message *lib.SimpleMessage, avoid string) {
	state.lock_peers.RLock()
	defer state.lock_peers.RUnlock()
	for addr, peer := range state.known_peers {
		if avoid != addr {
			gossiper.SendPacket(&lib.GossipPacket{Simple: message}, peer.Address, send_queue)
		}
	}
}

func client_handler(state State, server *lib.Gossiper, request lib.Packet) {
	packet := request.Content
	if state.simple && packet.Simple != nil {
		fmt.Println("CLIENT MESSAGE", packet.Simple.Contents)
		fmt.Println("PEERS", state)
		go state.broadcast(
			server,
			&lib.SimpleMessage{
				OriginalName:  server.Name,
				RelayPeerAddr: server.StringAddress,
				Contents:      packet.Simple.Contents},
			"")
	}

}

func server_handler(state State, server *lib.Gossiper, request lib.Packet) {
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
		// a status message can either be dispatched and use as an ack
		// or in the negative be used directly here
		if !state.dispatchStatusToPeer(packet.Status) {
			// anti entropy stuff
		}
	} else if packet.Rumor != nil {

	}

}

func main() {
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
	state := State{
		known_peers: make(map[string]*lib.Peer),
		db:          &db,
		simple:      *simple, lock_peers: &sync.RWMutex{}}

	for _, peer_addr := range peers_list {
		state.addPeer(peer_addr)
	}

	go gossiper.ReceiveLoop(msg_queue)
	go client_server.ReceiveLoop(client_queue)
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
