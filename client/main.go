package main

import (
	"flag"
	"github.com/dedis/protobuf"
	"github.com/poechsel/Peerster/lib"
	"net"
	"strings"
)

func main() {
	var port = flag.String("UIPort", "8080", "Port for the UI client")
	var dest = flag.String("dest", "", "destination for the private message")
	var file = flag.String("file", "", "file to be indexed by the gossiper, or filename of the requested file")
	var msg = flag.String("msg", "", "message to be sent")
	var request = flag.String("request", "", "request a chunk or metafile of this hash")
	var budget = flag.Int("budget", 0, "Budget for the file search")
	var keywords = flag.String("keywords", "", "Keywords to filter file with")
	flag.Parse()

	address := "127.0.0.1:" + *port
	udpAddr, err := lib.AddrOfString(address)
	lib.ExitIfError(err)
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	lib.ExitIfError(err)

	if *file != "" {
		if *request == "" {
			p := lib.NewDataRequest(*file, *dest, []byte{})
			gossip_packet :=
				&lib.GossipPacket{
					DataRequest: p}
			packetBytes, err := protobuf.Encode(gossip_packet)
			lib.ExitIfError(err)
			udpConn.Write(packetBytes)
		} else {
			p := lib.NewDataReply(*file, *dest, lib.UidToHash(*request), []byte{})
			gossip_packet :=
				&lib.GossipPacket{
					DataReply: p}
			packetBytes, err := protobuf.Encode(gossip_packet)
			lib.ExitIfError(err)
			udpConn.Write(packetBytes)
		}
	} else if *keywords != "" {
		p := lib.NewSearchRequest("", uint64(*budget), strings.Split(*keywords, ","))
		gossip_packet :=
			&lib.GossipPacket{
				SearchRequest: p}

		packetBytes, err := protobuf.Encode(gossip_packet)
		lib.ExitIfError(err)
		udpConn.Write(packetBytes)
	} else if *msg != "" {
		if *dest != "" {
			p := lib.NewPrivateMessage("client", *msg, *dest)
			gossip_packet :=
				&lib.GossipPacket{
					Private: &p}

			packetBytes, err := protobuf.Encode(gossip_packet)
			lib.ExitIfError(err)
			udpConn.Write(packetBytes)
		} else {
			gossip_packet :=
				&lib.GossipPacket{
					Simple: &lib.SimpleMessage{
						"client",
						address,
						*msg}}
			packetBytes, err := protobuf.Encode(gossip_packet)
			lib.ExitIfError(err)
			udpConn.Write(packetBytes)
		}
	}
}
