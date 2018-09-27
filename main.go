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

type State struct {
	mux         sync.RWMutex
	Known_peers map[string]*lib.Peer
	simple      bool
}

func (state *State) addPeer(address string) bool {
	state.mux.Lock()
	defer state.mux.Unlock()
	if _, ok := state.Known_peers[address]; address == "" || ok {
		return false
	} else {
		peer, err := lib.NewPeer(address)
		if err == nil {
			state.Known_peers[address] = peer
		}
		return true
	}
}

func (state *State) String() string {
	state.mux.RLock()
	defer state.mux.RUnlock()
	keys := make([]string, 0, len(state.Known_peers))
	for key := range state.Known_peers {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}

func (state *State) broadcast(gossiper *lib.Gossiper, message *lib.SimpleMessage, avoid string) {
	state.mux.RLock()
	defer state.mux.RUnlock()
	for addr, _ := range state.Known_peers {
		if avoid != addr {
			o, _ := lib.AddrOfString(addr)
			gossiper.SendGossipTo(&lib.GossipPacket{Simple: message}, o)
		}
	}
}

func receiver_loop(gossiper *lib.Gossiper, client_conn *lib.Gossiper, state *State) {
	for {
		packet, _, err := client_conn.ReceiveGossip()
		if err == nil {
			if state.simple && packet.Simple != nil {
				fmt.Println("CLIENT MESSAGE", packet.Simple.Contents)
				fmt.Println("PEERS", state)
				go state.broadcast(
					gossiper,
					&lib.SimpleMessage{
						OriginalName:  gossiper.Name,
						RelayPeerAddr: gossiper.StringAddress,
						Contents:      packet.Simple.Contents},
					"")
			}
		}
	}
}

func gossip_loop(gossiper *lib.Gossiper, state *State) {
	for {
		packet, source, err := gossiper.ReceiveGossip()
		source_string := lib.StringOfAddr(source)
		if err == nil {
			if state.simple && packet.Simple != nil {
				fmt.Println("SIMPLE MESSAGE", packet.Simple)
				state.addPeer(source_string)
				fmt.Println("PEERS", state)
				go state.broadcast(
					gossiper,
					&lib.SimpleMessage{
						OriginalName:  packet.Simple.OriginalName,
						RelayPeerAddr: gossiper.StringAddress,
						Contents:      packet.Simple.Contents},
					source_string)
			}
		}
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
	ignore(peers_list)

	gossiper, err := lib.NewGossiper(*gossip_addr, *gossip_name)
	ignore(err)

	client_server, err := lib.NewGossiper("127.0.0.1:"+*client_port, "client")
	ignore(err)

	state := State{Known_peers: make(map[string]*lib.Peer), simple: *simple}

	for _, peer_addr := range peers_list {
		state.addPeer(peer_addr)
	}

	go receiver_loop(gossiper, client_server, &state)
	go gossip_loop(gossiper, &state)

	// infinite loop
	for {
	}
}
