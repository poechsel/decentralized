package lib

import (
	//"log"
	"net"
	"sync"
)

/* Represent a peer. A peer can:
- have an address
- request 0, 1 or more status to be used as ack.
*/

type Peer struct {
	Address        *net.UDPAddr
	status_awaited int
	Status_channel chan *StatusPacket
	lock           *sync.Mutex
}

func NewPeer(address string) (*Peer, error) {
	a, err := AddrOfString(address)
	return &Peer{Address: a, lock: &sync.Mutex{},
		status_awaited: 0,
		Status_channel: make(chan *StatusPacket)}, err
}

/* Request an ack */
func (peer *Peer) RequestStatus() {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	peer.status_awaited += 1
}

/* Cancel the request for an ack */
func (peer *Peer) CancelRequestStatus() {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	if peer.status_awaited > 0 {
		peer.status_awaited -= 1
	}
}

/* If the peer waits for an ack, then he will be using the statuspacket
status and return true. Otherwise he will just return false */
func (peer *Peer) DispatchStatus(status *StatusPacket) bool {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	if peer.status_awaited > 0 {
		peer.status_awaited -= 1
		peer.Status_channel <- status
		return true
	} else {
		return false
	}
}
