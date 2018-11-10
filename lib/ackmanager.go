package lib

import ()

/* As we can't know if a channel is closed without reading
it, I'm using two channels. One is to send the ack, the
other is used as an indicator of the fact that the first
is closed */
type AckRequest struct {
	AckChannel chan interface{}
	isClosed   bool
}

func NewAckRequest() *AckRequest {
	x := AckRequest{AckChannel: make(chan interface{}), isClosed: false}
	return &x
}

func (ack *AckRequest) Close() {
	close(ack.AckChannel)
	ack.isClosed = true
}

func (ackr *AckRequest) SendAck(ack interface{}) bool {
	if !ackr.isClosed {
		ackr.AckChannel <- ack
		return true
	} else {
		return false
	}
}
