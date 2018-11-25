package lib

import (
	"sync"
)

/* As we can't know if a channel is closed without reading
it, I'm using two channels. One is to send the ack, the
other is used as an indicator of the fact that the first
is closed */
type AckRequest struct {
	AckChannel chan interface{}
	isClosed   bool
	lock       *sync.Mutex
}

func NewAckRequest() *AckRequest {
	x := AckRequest{
		AckChannel: make(chan interface{}),
		isClosed:   false,
		lock:       &sync.Mutex{}}
	return &x
}

func (ack *AckRequest) Close() {
	ack.lock.Lock()
	defer ack.lock.Unlock()
	close(ack.AckChannel)
	ack.isClosed = true
}

func (ackr *AckRequest) SendAck(ack interface{}) bool {
	ackr.lock.Lock()
	defer ackr.lock.Unlock()
	if !ackr.isClosed {
		ackr.AckChannel <- ack
		return true
	} else {
		return false
	}
}
