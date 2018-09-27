package lib

type Peer struct {
}

func NewPeer(address string) (*Peer, error) {
	return &Peer{}, nil
}
