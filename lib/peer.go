package lib

import "net"

type Peer struct {
	nb_ack_awaited int
	ack_channel    chan int
	Address        *net.UDPAddr
}

func NewPeer(address string) (*Peer, error) {
	a, err := AddrOfString(address)
	return &Peer{Address: a}, err
}
