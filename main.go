package main

import (
	"flag"
	"fmt"
	"github.com/poechsel/Peerster/lib"
	"net"
	"strings"
	"sync"
)

func ignore(x interface{}) {
	_ = x
}

type State struct {
	known_peers map[string]bool
	mux         sync.Mutex
}

func (state *State) addPeer(peer string) bool {
	if _, ok := state.known_peers[peer]; ok {
		return false
	} else {
		state.mux.Lock()
		state.known_peers[peer] = true
		state.mux.Unlock()
		return true
	}
}

func receiver_loop(client_conn *net.UDPConn, queue chan *lib.GossipPacket) {
	for {
		packet, err := lib.ReceiveGossip(client_conn)
		if err == nil {
			queue <- packet
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

	known_peers := make(map[string]bool)
	ignore(known_peers)
	peers_list := strings.Split(*peers_param, ",")
	ignore(peers_list)

	client_queue := make(chan *lib.GossipPacket)
	event_queue := make(chan *lib.GossipPacket)
	ignore(event_queue)

	gossiper, err := lib.NewGossiper(*gossip_addr, *gossip_name)
	ignore(err)
	ignore(gossiper)

	client_conn, _, err := lib.OpenPermanentConnection("127.0.0.1:" + *client_port)
	fmt.Println(err)

	go receiver_loop(client_conn, client_queue)

	for packet := range client_queue {
		fmt.Println(packet.Simple.Contents)
	}

	/*
		    for _, peer := range peers_list {
				peer_addr, err := lib.AddrOfString(peer)
				if err == nil {
					conn, err := net.ListenUDP("udp4", peer_addr)
					if err == nil {
						go func() {
							for {
								data, sender := conn.Receive()
								fmt.
							}
						}()
					}
				}
			}
	*/

	/*
		for _, event := range event_queue {

		}*/
}
