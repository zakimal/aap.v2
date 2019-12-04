package aap_v2

import (
	"bufio"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/zakimal/aap.v2/log"
	"github.com/zakimal/aap.v2/payload"
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type Peer struct {
	id uint64

	parent interface{} // Master or Worker

	conn net.Conn

	kill     chan *sync.WaitGroup
	killOnce uint32
}

func NewPeer(conn net.Conn, parent interface{}) *Peer {
	return &Peer{
		id:       0,
		parent:   parent,
		conn:     conn,
		kill:     make(chan *sync.WaitGroup, 2),
		killOnce: 0,
	}
}

func (p *Peer) init() {
	go p.messageReceiver()
}

func (p *Peer) messageReceiver() {
	reader := bufio.NewReader(p.conn)
	for {
		select {
		case wg := <-p.kill:
			wg.Done()
			return
		default:
		}
		size, err := binary.ReadUvarint(reader)
		if err != nil {
			p.DisconnectAsync()
			continue
		}
		buf := make([]byte, int(size))
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			p.DisconnectAsync()
			continue
		}
		opcode, msg, err := p.DecodeMessage(buf)
		if opcode == OpcodeNil || err != nil {
			p.DisconnectAsync()
			continue
		}

		var Q interface{}

		switch parent := p.parent.(type) {
		case *Master:
			Q, _ = parent.recvQueue.LoadOrStore(opcode, receiveHandle{
				hub:  make(chan Message),
				lock: make(chan struct{}, 1),
			})
		case *Worker:
			Q, _ = parent.recvQueue.LoadOrStore(opcode, receiveHandle{
				hub:  make(chan Message),
				lock: make(chan struct{}, 1),
			})
		default:
			panic("reachable")
		}

		recvQueue := Q.(receiveHandle)
		select {
		case recvQueue.hub <- msg:
			recvQueue.lock <- struct{}{}
			//switch parent := p.parent.(type) {
			//case *Worker:
			//	log.Print(parent)
			//	// TODO: adjust DS
			//}
			<-recvQueue.lock
		case <-time.After(3 * time.Second):
			p.DisconnectAsync()
			continue
		}
	}
}

func (p *Peer) DecodeMessage(buf []byte) (Opcode, Message, error) {
	reader := payload.NewReader(buf)
	opcode, err := reader.ReadByte()
	if err != nil {
		return OpcodeNil, nil, errors.Wrap(err, "failed to read opcode")
	}
	message, err := MessageFromOpcode(Opcode(opcode))
	if err != nil {
		return Opcode(opcode), nil, errors.Wrap(err, "opcode <-> message pairing not registered")
	}
	message, err = message.Read(reader)
	if err != nil {
		return Opcode(opcode), nil, errors.Wrap(err, "failed to read message contents")
	}
	return Opcode(opcode), message, nil
}

func (p *Peer) Disconnect() {
	if !atomic.CompareAndSwapUint32(&p.killOnce, 0, 1) {
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		p.kill <- &wg
	}
	if err := p.conn.Close(); err != nil {
		log.Info().Msg(errors.Wrapf(err, "got errors closing peer connection").Error())
	}
	wg.Wait()
	close(p.kill)
}

func (p *Peer) DisconnectAsync() <-chan struct{} {
	signal := make(chan struct{})
	if !atomic.CompareAndSwapUint32(&p.killOnce, 0, 1) {
		close(signal)
		return signal
	}
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		p.kill <- &wg
	}
	if err := p.conn.Close(); err != nil {
		log.Info().Msg(errors.Wrapf(err, "got errors closing peer connection").Error())
	}
	go func() {
		wg.Wait()
		close(p.kill)
		close(signal)
	}()
	return signal
}

func (p *Peer) ID() uint64 {
	return p.id
}

func (p *Peer) Parent() (reflect.Type, interface{}) {
	switch parent := p.parent.(type) {
	case *Master:
		return reflect.TypeOf(p.parent), parent
	case *Worker:
		return reflect.TypeOf(p.parent), parent
	default:
		panic("unreachable")
	}
}
