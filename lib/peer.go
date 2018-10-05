package lib

import (
	"net"
	"sync"
	"sync/atomic"
)

type Peer struct {
	Address        *net.UDPAddr
	status_awaited uint32
	Status_channel chan *StatusPacket
	lock           *sync.Mutex
}

func NewPeer(address string) (*Peer, error) {
	a, err := AddrOfString(address)
	return &Peer{Address: a}, err
}

func (peer *Peer) RequestStatus() {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	atomic.AddUint32(&peer.status_awaited, 1)
}

func (peer *Peer) CancelRequestStatus() {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	// substract 1, see doc
	if peer.status_awaited > 0 {
		atomic.AddUint32(&peer.status_awaited, ^uint32(0))
	}
}

func (peer *Peer) DispatchStatus(status *StatusPacket) bool {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	// substract 1, see doc
	if peer.status_awaited > 0 {
		atomic.AddUint32(&peer.status_awaited, ^uint32(0))
		peer.Status_channel <- status
		return true
	} else {
		return false
	}
}
