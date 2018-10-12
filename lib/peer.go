package lib

import (
	//"log"
	"net"
	"sync"
)

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

func (peer *Peer) RequestStatus() {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	//log.Println(peer.status_awaited)
	peer.status_awaited += 1
}

func (peer *Peer) CancelRequestStatus() {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	// substract 1, see doc
	//log.Println(peer.status_awaited)
	if peer.status_awaited > 0 {
		peer.status_awaited -= 1
	}
}

func (peer *Peer) DispatchStatus(status *StatusPacket) bool {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	//log.Println(peer.status_awaited)
	// substract 1, see doc
	if peer.status_awaited > 0 {
		peer.status_awaited -= 1
		peer.Status_channel <- status
		return true
	} else {
		return false
	}
}
