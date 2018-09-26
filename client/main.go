package main

import (
	"flag"
	"github.com/poechsel/Peerster/lib"
)

func main() {
	var port = flag.String("UIPort", "8080", "Port for the UI client")
	var msg = flag.String("msg", "", "message to be sent")
	flag.Parse()

	address := "127.0.0.1:" + *port
	peer, _ := lib.NewPeer(address)
	gossip_packet :=
		&lib.GossipPacket{
			&lib.SimpleMessage{
				"client",
				address,
				*msg}}
	peer.SendGossip(gossip_packet)
}
