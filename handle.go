package aap_v2

type sendHandle struct {
	to      *Peer
	payload []byte
	result  chan error
}

type receiveHandle struct {
	hub  chan Message
	lock chan struct{}
}
