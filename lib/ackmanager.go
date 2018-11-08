package lib

import ()

/* As we can't know if a channel is closed without reading
it, I'm using two channels. One is to send the ack, the
other is used as an indicator of the fact that the first
is closed */
type AckRequest struct {
	AckChannel chan interface{}
	isClosed   chan bool
}

func NewAckRequest() *AckRequest {
	return &AckRequest{AckChannel: make(chan interface{}), isClosed: make(chan bool)}
}

func (ack *AckRequest) Close() {
	close(ack.AckChannel)
	close(ack.isClosed)
}

func (ackr *AckRequest) SendAck(ack interface{}) bool {
	if _, ok := <-ackr.isClosed; ok {
		ackr.AckChannel <- ack
		return true
	} else {
		return false
	}
}
