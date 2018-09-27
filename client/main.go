package main

import (
	"flag"
	"github.com/dedis/protobuf"
	"github.com/poechsel/Peerster/lib"
	"net"
)

func main() {
	var port = flag.String("UIPort", "8080", "Port for the UI client")
	var msg = flag.String("msg", "", "message to be sent")
	flag.Parse()

	address := "127.0.0.1:" + *port
	udpAddr, err := lib.AddrOfString(address)
	lib.ExitIfError(err)
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	lib.ExitIfError(err)
	gossip_packet :=
		&lib.GossipPacket{
			&lib.SimpleMessage{
				"client",
				address,
				*msg}}
	packetBytes, err := protobuf.Encode(gossip_packet)
	lib.ExitIfError(err)
	udpConn.Write(packetBytes)
}
