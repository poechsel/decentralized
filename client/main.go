package main

import (
	"flag"
	"github.com/poechsel/Peerster/lib"
)

func main() {
	var port = flag.String("UIPort", "8080", "Port for the UI client")
	var msg = flag.String("msg", "", "message to be sent")
	flag.Parse()
	conn, _, _ := lib.OpenWriteConnection("127.0.0.1:" + *port)
	gossip_packet := &lib.GossipPacket{&lib.SimpleMessage{"a", "b", *msg}}
	lib.SendGossip(conn, gossip_packet)
}
