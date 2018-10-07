package lib

import (
	"fmt"
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
	return &Peer{Address: a, lock: &sync.Mutex{},
		Status_channel: make(chan *StatusPacket)}, err
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
	fmt.Println("waiting for lock")
	peer.lock.Lock()
	fmt.Println("YES")
	defer peer.lock.Unlock()
	// substract 1, see doc
	fmt.Println("YES")
	if peer.status_awaited > 0 {
		fmt.Println("YES")
		atomic.AddUint32(&peer.status_awaited, ^uint32(0))
		fmt.Println("trying to send")
		peer.Status_channel <- status
		fmt.Println("send success")
		return true
	} else {
		return false
	}
}
