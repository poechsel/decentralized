package main

import (
	"flag"
	"github.com/poechsel/Peerster/lib"
)

func main() {
	var port = flag.String("UIPort", "8080", "Port for the UI client")
	var msg = flag.String("msg", "", "message to be sent")
	flag.Parse()

	peer, _ := lib.NewPeer("127.0.0.1:" + *port)
	gossip_packet :=
		&lib.GossipPacket{
			&lib.SimpleMessage{
				"client",
				peer.CanonicalAddress,
				*msg}}
	peer.SendGossip(gossip_packet)
}
