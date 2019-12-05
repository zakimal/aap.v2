package aap_v2

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zakimal/aap.v2/log"
	"github.com/zakimal/aap.v2/payload"
	"github.com/zakimal/aap.v2/transport"
	"io"
	"net"
	"os"
	"strconv"
)

type Master struct {
	id uint64

	transport transport.Transport
	listener  net.Listener

	host string
	port uint16

	workers map[uint64]*Peer

	sendQueue chan sendHandle
	recvQueue map[Opcode]receiveHandle

	inactiveMap  map[uint64]bool
	terminateMap map[uint64]bool
	assembleMap  map[uint64]bool

	kill     chan chan struct{}
	killOnce uint32
}

func NewMaster(id uint64) (*Master, error) {
	config, err := os.Open(fmt.Sprintf("config/workers/%d.csv", id))
	if err != nil {
		panic(err)
	}
	defer config.Close()

	configReader := csv.NewReader(config)
	record, err := configReader.Read()
	if err != nil {
		panic(err)
	}
	host := record[0]
	port, err := strconv.ParseInt(record[1], 10, 16)
	if err != nil {
		panic(err)
	}

	tcp := transport.NewTCP()
	listener, err := tcp.Listen(host, uint16(port))
	if err != nil {
		return nil, errors.Errorf("failed to create listener for peer on %s:%d", host, port)
	}

	master := Master{
		id:           id,
		transport:    tcp,
		listener:     listener,
		host:         host,
		port:         uint16(port),
		workers:      make(map[uint64]*Peer),
		sendQueue:    make(chan sendHandle, 128),
		recvQueue:    make(map[Opcode]receiveHandle),
		inactiveMap:  make(map[uint64]bool),
		terminateMap: make(map[uint64]bool),
		assembleMap:  make(map[uint64]bool),
		kill:         make(chan chan struct{}, 1),
		killOnce:     0,
	}

	master.init()

	master.recvQueue[opcodeHelloWorker] = receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodeHelloMaster] = receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodePEvalRequest] = receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodePEvalResponse] =  receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodeIncEvalUpdate] =  receiveHandle{
		hub: make(chan Message, 128),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodeNotifyInactive] =  receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodeTerminateRequest] =  receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodeTerminateACK] =  receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodeTerminateNACK] =  receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodeAssembleRequest] =  receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}
	master.recvQueue[opcodeAssembleResponse] =  receiveHandle{
		hub: make(chan Message),
		lock:make(chan struct {}, 1),
	}

	return &master, nil
}

func (m *Master) init() {
	opcodeHelloWorker = RegisterMessage(NextAvailableOpcode(), (*MessageHelloWorker)(nil))
	opcodeHelloMaster = RegisterMessage(NextAvailableOpcode(), (*MessageHelloMaster)(nil))
	opcodePEvalRequest = RegisterMessage(NextAvailableOpcode(), (*MessagePEvalRequest)(nil))
	opcodePEvalResponse = RegisterMessage(NextAvailableOpcode(), (*MessagePEvalResponse)(nil))
	opcodeIncEvalUpdate = RegisterMessage(NextAvailableOpcode(), (*MessageIncEvalUpdate)(nil))
	opcodeNotifyInactive = RegisterMessage(NextAvailableOpcode(), (*MessageNotifyInactive)(nil))
	opcodeTerminateRequest = RegisterMessage(NextAvailableOpcode(), (*MessageTerminateRequest)(nil))
	opcodeTerminateACK = RegisterMessage(NextAvailableOpcode(), (*MessageTerminateACK)(nil))
	opcodeTerminateNACK = RegisterMessage(NextAvailableOpcode(), (*MessageTerminateNACK)(nil))
	opcodeAssembleRequest = RegisterMessage(NextAvailableOpcode(), (*MessageAssembleRequest)(nil))
	opcodeAssembleResponse = RegisterMessage(NextAvailableOpcode(), (*MessageAssembleResponse)(nil))

	log.Info().Msgf("opcodeHelloWorker=%d", opcodeHelloWorker)
	log.Info().Msgf("opcodeHelloMaster=%d", opcodeHelloMaster)
	log.Info().Msgf("opcodePEvalRequest=%d", opcodePEvalRequest)
	log.Info().Msgf("opcodePEvalResponse=%d", opcodePEvalResponse)
	log.Info().Msgf("opcodeIncEvalUpdate=%d", opcodeIncEvalUpdate)
	log.Info().Msgf("opcodeNotifyInactive=%d", opcodeNotifyInactive)
	log.Info().Msgf("opcodeTerminateRequest=%d", opcodeTerminateRequest)
	log.Info().Msgf("opcodeTerminateACK=%d", opcodeTerminateACK)
	log.Info().Msgf("opcodeTerminateNACK=%d", opcodeTerminateNACK)
	log.Info().Msgf("opcodeAssembleRequest=%d", opcodeAssembleRequest)
	log.Info().Msgf("opcodeAssembleResponse=%d", opcodeAssembleResponse)
	go m.messageSender()
}

func (m *Master) messageSender() {
	for {
		var cmd sendHandle
		select {
		case cmd = <-m.sendQueue:
		}
		to := cmd.to
		payload := cmd.payload
		size := len(payload)
		buf := make([]byte, binary.MaxVarintLen64)
		prepend := binary.PutUvarint(buf[:], uint64(size))
		buf = append(buf[:prepend], payload[:]...)
		copied, err := io.Copy(to.conn, bytes.NewReader(buf))
		if copied != int64(size+prepend) {
			if cmd.result != nil {
				cmd.result <- errors.Errorf(
					"only written %d bytes when expected to write %d bytes to peer\n",
					copied, size+prepend)
				close(cmd.result)
			}
			continue
		}
		if err != nil {
			if cmd.result != nil {
				cmd.result <- errors.Wrap(err, "failed to send message to peer")
				close(cmd.result)
			}
			continue
		}
		if cmd.result != nil {
			cmd.result <- nil
			close(cmd.result)
		}
	}
}

func (m *Master) SendMessage(to *Peer, message Message) error {
	payload, err := m.EncodeMessage(message)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize message contents to be sent to a peer")
	}
	cmd := sendHandle{
		to:      to,
		payload: payload,
		result:  make(chan error, 1),
	}
	select {
	case m.sendQueue <- cmd:
	}
	select {
	case err = <-cmd.result:
		return err
	}
}

func (m *Master) SendMessageAsync(to *Peer, message Message) <-chan error {
	result := make(chan error, 1)
	payload, err := m.EncodeMessage(message)
	if err != nil {
		result <- errors.Wrap(err, "failed to serialize message contents to be sent to a peer")
		return result
	}
	cmd := sendHandle{
		to:      to,
		payload: payload,
		result:  result,
	}
	select {
	case m.sendQueue <- cmd:
	}
	return result
}

func (m *Master) EncodeMessage(message Message) ([]byte, error) {
	opcode, err := OpcodeFromMessage(message)
	if err != nil {
		return nil, errors.Wrap(err, "could not find opcode registered for message")
	}
	var buf bytes.Buffer
	_, err = buf.Write(payload.NewWriter(nil).WriteByte(byte(opcode)).Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize message opcode")
	}
	_, err = buf.Write(message.Write())
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize and write message contents")
	}
	return buf.Bytes(), nil
}

func (m *Master) Receive(opcode Opcode) <-chan Message {
	c, _ := m.recvQueue[opcode]
	return c.hub
}

func (m *Master) Listen() {
	for {
		select {
		case signal := <-m.kill:
			close(signal)
			return
		default:
		}
		conn, err := m.listener.Accept()
		if err != nil {
			continue
		}
		log.Info().Msgf("accept connection from %s", conn.LocalAddr().String())
		peer := NewPeer(conn, m)
		peer.init()
		if err := m.SendMessage(peer, MessageHelloMaster{from: m.id}); err != nil {
			panic(err)
		}
		select {
		case msg := <-m.Receive(opcodeHelloMaster):
			pid := msg.(MessageHelloMaster).from
			peer.id = pid
			m.workers[pid] = peer
		}
	}
}

func (m *Master) Dial(address string) (*Peer, error) {
	conn, err := m.transport.Dial(address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to peer %s", conn)
	}
	peer := NewPeer(conn, m)
	peer.init()
	if err := m.SendMessage(peer, MessageHelloMaster{from: m.id}); err != nil {
		panic(err)
	}
	select {
	case msg := <-m.Receive(opcodeHelloMaster):
		pid := msg.(MessageHelloMaster).from
		peer.id = pid
		m.workers[pid] = peer
		m.inactiveMap[pid] = false
		m.terminateMap[pid] = false
		m.assembleMap[pid] = false
	}
	return peer, nil
}

func (m *Master) Disconnect(peer *Peer) {
	id := peer.id
	delete(m.workers, id)
	peer.Disconnect()
}

func (m *Master) DisconnectAsync(peer *Peer) <-chan struct{} {
	id := peer.id
	delete(m.workers, id)
	return peer.DisconnectAsync()
}

func (m *Master) Run() {
	for _, worker := range m.workers {
		if err := m.SendMessage(worker, MessagePEvalRequest{from: m.id}); err != nil {
			panic(err)
		}
	}
	log.Info().Msg("Broadcast PEval Request")

WaitInactive:
	for {
		select {
		case msg := <-m.Receive(opcodeNotifyInactive):
			wid := msg.(MessageNotifyInactive).from
			log.Info().Msgf("Worker %d notifies inactive", wid)
			m.inactiveMap[wid] = true
			flag := true
			for _, st := range m.inactiveMap {
				flag = flag && st
			}
			if flag {
				for _, w := range m.workers {
					if err := m.SendMessage(w, MessageTerminateRequest{from: m.id}); err != nil {
						panic(err)
					}
				}
				log.Info().Msg("Broadcast Terminate")
				goto WaitTerminateAck
			}
		}
	}

WaitTerminateAck:
	for {
		select {
		case msg := <-m.Receive(opcodeTerminateACK):
			wid := msg.(MessageTerminateACK).from
			log.Info().Msgf("Worker %d is inactive", wid)
			m.terminateMap[wid] = true
			flag := true
			for _, st := range m.terminateMap {
				flag = flag && st
			}
			if flag {
				goto WaitAssembleResponse
			}
		case msg := <-m.Receive(opcodeTerminateNACK):
			wid := msg.(MessageTerminateNACK).from
			log.Info().Msgf("Worker %d is active", wid)
			m.inactiveMap[wid] = false
			goto WaitInactive
		}
	}

WaitAssembleResponse:
	for _, worker := range m.workers {
		if err := m.SendMessage(worker, MessageAssembleRequest{from: m.id}); err != nil {
			panic(err)
		}
	}
	for {
		select {
		case msg := <-m.Receive(opcodeAssembleResponse):
			wid := msg.(MessageAssembleResponse).from
			result := msg.(MessageAssembleResponse).result
			m.assembleMap[wid] = true
			log.Info().Msgf("Worker %d returns partial result: %+v", wid, result)
			flag := true
			for _, st := range m.terminateMap {
				flag = flag && st
			}
			if flag {
				goto Terminate
			}
		}
	}

Terminate:
	log.Info().Msg("Bye ;)")
}

func (m *Master) ID() uint64 {
	return m.id
}

func (m *Master) Host() string {
	return m.host
}

func (m *Master) Port() uint16 {
	return m.port
}

func (m *Master) Address() string {
	return fmt.Sprintf("%s:%d", m.host, m.port)
}

func (m *Master) Workers() map[uint64]*Peer {
	return m.workers
}
