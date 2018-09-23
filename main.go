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
	Known_peers map[string]*lib.Peer
	Mux         sync.Mutex
}

func (state *State) addPeer(address string) bool {
	if _, ok := state.Known_peers[address]; ok {
		return false
	} else {
		state.Mux.Lock()
		if peer, err := lib.NewPeer(address); err != nil {
			state.Known_peers[peer.CanonicalAddress] = peer
		}
		state.Mux.Unlock()
		return true
	}
}

func broadcast(message *lib.SimpleMessage, state *State, avoid *string) {
	// should think about the lock
	for addr, peer := range state.Known_peers {
		if avoid != nil && *avoid == addr {
			peer.SendGossip(&lib.GossipPacket{Simple: message})
		}
	}
}

func receiver_loop(gossiper *lib.Gossiper, client_conn *lib.Gossiper, state *State) {
	for {
		packet, err := client_conn.ReceiveGossip()
		if err == nil {
			if packet.Simple != nil {
				fmt.Println(packet.Simple.Contents)
				go broadcast(
					&lib.SimpleMessage{
						OriginalName:  gossiper.Name,
						RelayPeerAddr: gossiper.CanonicalAddress,
						Contents:      packet.Simple.Contents},
					state,
					nil)
			}
		}
	}
}

func gossip_loop(gossiper *lib.Gossiper, state *State) {
	for {
		packet, err := gossiper.ReceiveGossip()
		if err == nil {
			if packet.Simple != nil {
				fmt.Println(packet.Simple.Contents)
				go state.addPeer(packet.Simple.RelayPeerAddr)
				go broadcast(
					&lib.SimpleMessage{
						OriginalName:  packet.Simple.OriginalName,
						RelayPeerAddr: gossiper.CanonicalAddress,
						Contents:      packet.Simple.Contents},
					state,
					&packet.Simple.RelayPeerAddr)
			}
		}
	}
}

func main() {
	client_port := flag.String("UIPort", "8080", "Port for the UI client")
	gossip_addr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	gossip_name := flag.String("name", "", "name of the gossiper")
	peers_param := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	var _ = flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	flag.Parse()

	peers_list := strings.Split(*peers_param, ",")
	ignore(peers_list)

	gossiper, err := lib.NewGossiper(*gossip_addr, *gossip_name)
	ignore(err)

	client_server, err := lib.NewGossiper("127.0.0.1:"+*client_port, "client")
	ignore(err)

	state := State{}

	for _, peer_addr := range peers_list {
		if peer, err := lib.NewPeer(peer_addr); err != nil {
			state.Known_peers[peer.CanonicalAddress] = peer
		}
	}

	go receiver_loop(gossiper, client_server, &state)
	go gossip_loop(gossiper, &state)

	// infinite loop
	for {
	}
}
