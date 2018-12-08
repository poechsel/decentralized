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
var server_queue = make(lib.NetChannel)

type PeerId struct {
	Address string
	Name    string
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	/*
		BlockChain := lib.NewBlockChain()
		conv := make(map[[32]byte]string)

		conv[[32]byte{}] = "0"

		a := lib.NewBlock([32]byte{}, [32]byte{}, []lib.TxPublish{})
		conv[a.Hash()] = "a"
		BlockChain.AppendBlock(conv, a)
		b := lib.NewBlock(a.Hash(), [32]byte{}, []lib.TxPublish{})
		conv[b.Hash()] = "b"
		BlockChain.AppendBlock(conv, b)
		c := lib.NewBlock(b.Hash(), [32]byte{}, []lib.TxPublish{})
		conv[c.Hash()] = "c"

		d := lib.NewBlock(c.Hash(), [32]byte{}, []lib.TxPublish{})
		conv[d.Hash()] = "d"
		BlockChain.AppendBlock(conv, d)
		e := lib.NewBlock(d.Hash(), [32]byte{}, []lib.TxPublish{})
		conv[e.Hash()] = "e"
		BlockChain.AppendBlock(conv, e)

		x := [32]byte{}
		x[0] = 12
		f := lib.NewBlock(c.Hash(), x, []lib.TxPublish{})
		conv[f.Hash()] = "f"
		BlockChain.AppendBlock(conv, f)
		g := lib.NewBlock(f.Hash(), [32]byte{}, []lib.TxPublish{})
		conv[g.Hash()] = "g"
		BlockChain.AppendBlock(conv, g)
		h := lib.NewBlock(g.Hash(), [32]byte{}, []lib.TxPublish{})
		conv[h.Hash()] = "h"
		BlockChain.AppendBlock(conv, h)
		BlockChain.Dump(conv)
		BlockChain.AppendBlock(conv, c)
		BlockChain.Dump(conv)

		return
	*/
	/* Parse the command line */
	client_port := flag.String("UIPort", "8080", "Port for the UI client")
	gossip_addr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	gossip_name := flag.String("name", "123456789", "name of the gossiper")
	peers_param := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	rtimer := flag.Int("rtimer", 0, "route rumors sending period in seconds, 0 to disable sending of route rumors")
	var simple = flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	flag.Parse()
	peers_list := strings.Split(*peers_param, ",")

	lib.InitializeTempDir(*gossip_name)
	/* create the current gossiper */
	gossiper, err := lib.NewGossiper(*gossip_addr, *gossip_name, *simple, *rtimer)
	fmt.Println("LISTENING ON: ", *gossip_addr)
	lib.ExitIfError(err)
	state := lib.NewState()
	state.UpdateRoutingTable(gossiper.Name, gossiper.Address.String())

	client_url := "127.0.0.1:" + *client_port
	/* If the UIPort is 8080, it means that we want to interact with
	the server using the web frontend. This will launch the webserver */
	if *client_port == "8080" {
		web := lib.NewWebServer(state, gossiper, client_url)
		state.AddNewMessageCallback(web.AddMessageChannel)
		state.AddNewPrivateMessageCallback(web.AddPrivateMessageChannel)
		state.AddNewSearchResultCallback(web.AddSearchResultChannel)
		state.AddNewPeerCallback(web.AddPeerChannel)
		go web.Start()
	} else {
		/* otherwise, we connect to the client and we wait to receive
		corresponding messages */
		client_server, err := lib.NewGossiper(client_url, "client", *simple, 0)
		lib.ExitIfError(err)
		go client_server.ReceiveLoop(client_queue)
	}

	/* Add the peers given as parameters */
	for _, peer_addr := range peers_list {
		state.AddPeer(peer_addr)
	}

	/* Listen for incoming messages */
	go gossiper.ReceiveLoop(server_queue)

	/* Launch antientropy if needed */
	if !gossiper.SimpleMode {
		gossiper.AntiEntropy(state)
	}

	go gossiper.RefreshRouteLoop(state)
	go gossiper.ListenBlockChainEvents(state)
	go state.BlockChain.Work()

	/* loop on incoming messages */
	for {
		select {
		case request := <-client_queue:
			go gossiper.ClientHandler(state, request)

		case request := <-server_queue:
			go gossiper.ServerHandler(state, request)

			/* this case should not be useful.
			However, if not putting every message send in this channel,
			when using the naive gossiper in Q1, some messages will not be send */
		case write := <-lib.Send_queue:
			lib.SendPacket(gossiper.Conn, write)
		}
	}
}
