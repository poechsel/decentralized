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
	mux         sync.Mutex
	Known_peers map[string]*lib.Peer
}

func (state *State) addPeer(address string) bool {
	fmt.Println("trying to add", address)
	if address == "" {
		return false
	}
	if _, ok := state.Known_peers[address]; ok {
		return false
	} else {
		state.Mux.Lock()
		/* Why rechecking for the existence of the key here as we
		already did it three lines ago ?
		For performances reasons. We don't want to block the
		program each time we try to add a Peer.
		The first check allows us to prune quickly. However, two
		routines might want to add the same peer at the same time.
		This will thus allocate twice a connection, which we don't want.
		Thus, once we are inside the lock, we are checking again
		*/
		fmt.Println("trying to add", address)
		if _, ok := state.Known_peers[address]; !ok {
			peer, err := lib.NewPeer(address)
			fmt.Println(peer, err)
			if err == nil {
				fmt.Println("trying to add", address)
				fmt.Println(state.Known_peers, peer, peer.CanonicalAddress)
				state.Known_peers[peer.CanonicalAddress] = peer
				fmt.Println("done")
			}
		}
		state.Mux.Unlock()
		return true
	}
}

func broadcast(message *lib.SimpleMessage, state *State, avoid *string) {
	/* No need to lock here. In the most extreme case we
	will send the message to more persons */
	// should think about the lock
	fmt.Println("broadcasting")
	for addr, peer := range state.Known_peers {
		if avoid == nil || *avoid != addr {
			fmt.Println("sending to", peer.CanonicalAddress)
			peer.SendGossip(&lib.GossipPacket{Simple: message})
		}
	}
}

func receiver_loop(gossiper *lib.Gossiper, client_conn *lib.Gossiper, state *State) {
	for {
		packet, err := client_conn.ReceiveGossip()
		if err == nil {
			if packet.Simple != nil {
				fmt.Println("CLIENT MESSAGE", packet.Simple.Contents)
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
				fmt.Println(
					"SIMPLE MESSAGE",
					packet.Simple)
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

	state := State{Known_peers: make(map[string]*lib.Peer)}

	for _, peer_addr := range peers_list {
		state.addPeer(peer_addr)
	}

	go receiver_loop(gossiper, client_server, &state)
	go gossip_loop(gossiper, &state)

	// infinite loop
	for {
	}
}
