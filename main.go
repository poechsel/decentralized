package main

import (
	"flag"
	"fmt"
	"github.com/poechsel/Peerster/lib"
	"math/rand"
	"strings"
	"time"
)

var client_queue = make(lib.NetChannel)
var msg_queue = make(lib.NetChannel)

type PeerId struct {
	Address string
	Name    string
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

	gossiper, err := lib.NewGossiper(*gossip_addr, *gossip_name, *simple)
	fmt.Println("LISTENING ON: ", *gossip_addr)
	lib.ExitIfError(err)

	client_url := "127.0.0.1:" + *client_port

	state := lib.NewState()
	if *client_port != "8080" {
		client_server, err := lib.NewGossiper(client_url, "client", *simple)
		lib.ExitIfError(err)
		go client_server.ReceiveLoop(client_queue)
	} else {
		web := lib.NewWebServer(state, gossiper, client_url, *gossip_name)
		state.AddNewMessageCallback(web.AddMessageChannel)
		state.AddNewPeerCallback(web.AddPeerChannel)
		go web.Start()
	}

	for _, peer_addr := range peers_list {
		state.AddPeer(peer_addr)
	}

	go gossiper.ReceiveLoop(msg_queue)

	if !gossiper.SimpleMode {
		gossiper.AntiEntropy(state)
	}
	// infinite loop
	for {
		select {
		case request := <-client_queue:
			go gossiper.ClientHandler(state, request)

		case request := <-msg_queue:
			go gossiper.ServerHandler(state, request)

		case write := <-lib.Send_queue:
			// should probably be removed, but ain't nobody got time for that
			lib.SendPacket(gossiper.Conn, write)
		}
	}
}
