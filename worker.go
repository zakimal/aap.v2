package aap_v2

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zakimal/aap.v2/graph"
	gio "github.com/zakimal/aap.v2/graph/io"
	"github.com/zakimal/aap.v2/graph/path"
	"github.com/zakimal/aap.v2/log"
	"github.com/zakimal/aap.v2/payload"
	"github.com/zakimal/aap.v2/transport"
	"github.com/zakimal/catalogue/aap/graph/simple"
	"io"
	"math"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Worker struct {
	id uint64

	transport transport.Transport
	listener  net.Listener

	host string
	port uint16

	peers map[uint64]*Peer

	sendQueue chan sendHandle
	recvQueue sync.Map

	round uint64

	g graph.Graph

	shortest path.Shortest

	// fit, fof, fot, fif
	// fif = fringe, inward, from
	// fit = fringe, outward, to
	// fof = fringe, inward, from
	// fot = fringe, outward, to
	fif map[graph.Vertex]uint64
	fit map[graph.Vertex]uint64
	fof map[graph.Vertex]uint64
	fot map[graph.Vertex]uint64

	master *Peer

	// TODO: need to think twice
	isInactive uint32 // 0: active, 1: inactive

	// TODO
	DS uint64 // Delay Stretch

	kill     chan chan struct{}
	killOnce uint32
}

func NewWorker(id uint64) (*Worker, error) {
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

	g, fit, fif, fot, fof := gio.BuildEdgeCutWeightedDirectedGraph(id)

	worker := Worker{
		id:         id,
		transport:  tcp,
		listener:   listener,
		host:       host,
		port:       uint16(port),
		peers:      make(map[uint64]*Peer),
		sendQueue:  make(chan sendHandle, 128),
		recvQueue:  sync.Map{},
		round:      0,
		g:          g,
		shortest:   path.Shortest{},
		fif:        fif,
		fit:        fit,
		fof:        fof,
		fot:        fot,
		master:     nil,
		isInactive: 0,
		kill:       make(chan chan struct{}, 1),
		killOnce:   0,
	}

	worker.init()

	return &worker, nil
}

func (w *Worker) init() {
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
	go w.messageSender()
}

func (w *Worker) messageSender() {
	for {
		var cmd sendHandle
		select {
		case cmd = <-w.sendQueue:
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

func (w *Worker) SendMessage(to *Peer, message Message) error {
	payload, err := w.EncodeMessage(message)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize message contents to be sent to a peer")
	}
	cmd := sendHandle{
		to:      to,
		payload: payload,
		result:  make(chan error, 1),
	}
	select {
	case w.sendQueue <- cmd:
	}
	select {
	case err = <-cmd.result:
		return err
	}
}

func (w *Worker) SendMessageAsync(to *Peer, message Message) <-chan error {
	result := make(chan error, 1)
	payload, err := w.EncodeMessage(message)
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
	case w.sendQueue <- cmd:
	}
	return result
}

func (w *Worker) EncodeMessage(message Message) ([]byte, error) {
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

func (w *Worker) Receive(opcode Opcode) <-chan Message {
	c, _ := w.recvQueue.LoadOrStore(opcode, receiveHandle{
		hub:  make(chan Message),
		lock: make(chan struct{}, 1),
	})
	return c.(receiveHandle).hub
}

func (w *Worker) Listen() {
	for {
		select {
		case signal := <-w.kill:
			close(signal)
			return
		default:
		}
		conn, err := w.listener.Accept()
		if err != nil {
			continue
		}
		log.Info().Msgf("accept connection from %s", conn.LocalAddr().String())
		peer := NewPeer(conn, w)
		peer.init()
		select {
		case msg := <-w.Receive(opcodeHelloWorker):
			pid := msg.(MessageHelloWorker).from
			peer.id = pid
			w.peers[pid] = peer
			if err := w.SendMessage(peer, MessageHelloWorker{from: w.id}); err != nil {
				panic(err)
			}
		case msg := <-w.Receive(opcodeHelloMaster):
			from := msg.(MessageHelloMaster).from
			w.master = peer
			w.master.id = from
			if err := w.SendMessage(w.master, MessageHelloMaster{from: w.id}); err != nil {
				panic(err)
			}
		}
	}
}

func (w *Worker) Dial(address string) (*Peer, error) {
	conn, err := w.transport.Dial(address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to peer %s", conn)
	}
	peer := NewPeer(conn, w)
	peer.init()
	if err := w.SendMessage(peer, MessageHelloWorker{from: w.id}); err != nil {
		panic(err)
	}
	select {
	case msg := <-w.Receive(opcodeHelloWorker):
		pid := msg.(MessageHelloWorker).from
		peer.id = pid
		w.peers[pid] = peer
	}
	return peer, nil
}

func (w *Worker) Disconnect(peer *Peer) {
	id := peer.id
	delete(w.peers, id)
	peer.Disconnect()
}

func (w *Worker) DisconnectAsync(peer *Peer) <-chan struct{} {
	id := peer.id
	delete(w.peers, id)
	return peer.DisconnectAsync()
}

func (w *Worker) Run() {
	msg := <-w.Receive(opcodePEvalRequest)
	mid := msg.(MessagePEvalRequest).from
	log.Info().Msgf("Receive PEval Request from master %d", mid)

	if w.master.id != mid {
		panic(fmt.Sprintf("PEval Request from wrong master: expected: %d, got: %d", w.master.id, mid))
	}

	log.Info().Msg("----- START PEVAL -----")
	var updateMap map[int64]float64
	w.shortest, updateMap = path.PEvalDijkstraFrom(w.g.Vertex(0), w.g)
	log.Info().Msgf("PEval: %+v, update=%+v", w.shortest, updateMap)

	for vid, dist := range updateMap {
		peer := w.peers[w.fot[w.g.Vertex(vid)]]
		if err := w.SendMessage(peer, MessageIncEvalUpdate{
			from:  mid,
			round: w.round,
			vid:   vid,
			dist:  dist,
		}); err != nil {
			panic(err)
		}
		log.Info().Msgf("Send update message to worker %d", peer.id)
	}

	//if len(w.Receive(opcodeIncEvalUpdate)) == 0 {
	//	w.isInactive = 1
	//	w.round += 1
	//	goto WaitTerminateRequest
	//}

IncrementalEvaluation:
	for range time.Tick(1 * time.Second) {
		log.Info().Msg("----- START INCEVAL -----")
		length := len(w.Receive(opcodeIncEvalUpdate))
		log.Info().Msgf("Receive %d updates", length)
		updateMap := make(map[int64]float64)
		min := uint64(math.MaxUint64)
		max := uint64(0)
		// TODO: FIX HERE
		// IncEvalUpdateを受け取ったらlen()が非ゼロになると思っていたのだけれど
		// メッセージは送られているはずなのに長さがゼロのままなのでだめ
		for len(w.Receive(opcodeIncEvalUpdate)) > 0 {
			msg := <-w.Receive(opcodeIncEvalUpdate)
			from := msg.(MessageIncEvalUpdate).from
			round := msg.(MessageIncEvalUpdate).round
			vid := msg.(MessageIncEvalUpdate).vid
			dist := msg.(MessageIncEvalUpdate).dist
			log.Info().Msgf("Receive update message: from=%d, round=%d, vid=%d, dist=%f",
				from, round, vid, dist)
			if min > round {
				min = round
			}
			if max < round {
				max = round
			}
			if _, exists := updateMap[vid]; exists {
				if dist < updateMap[vid] {
					updateMap[vid] = dist
				}
			} else {
				updateMap[vid] = dist
			}
		}
		log.Info().Msgf("max: %d, min: %d, now: %d", max, min, w.round)
		updateMap = path.IncEvalDijkstraFrom(updateMap, &w.shortest, simple.NewVertex(0), w.g)
		log.Info().Msgf("IncEval #%d: %+v, update=%+v", w.round, w.shortest, updateMap)
		for vid, dist := range updateMap {
			peer := w.peers[w.fot[w.g.Vertex(vid)]]
			if err := w.SendMessage(peer, MessageIncEvalUpdate{
				from:  mid,
				round: w.round,
				vid:   vid,
				dist:  dist,
			}); err != nil {
				panic(err)
			}
		}
		// TODO: adjust DS
		w.round += 1
		if len(w.Receive(opcodeIncEvalUpdate)) == 0 {
			w.isInactive = 1
			break IncrementalEvaluation
		}
	}
	//for {
	//	ticker := time.NewTicker(time.Second)
	//	select {
	//	case <-ticker.C: // elapsed DS time
	//		log.Info().Msg("----- START INCEVAL -----")
	//		length := len(w.Receive(opcodeIncEvalUpdate))
	//		log.Info().Msgf("Receive %d updates", length)
	//		updateMap := make(map[int64]float64)
	//		min := uint64(math.MaxUint64)
	//		max := uint64(0)
	//		for msg := range w.Receive(opcodeIncEvalUpdate) {
	//			from := msg.(MessageIncEvalUpdate).from
	//			round := msg.(MessageIncEvalUpdate).round
	//			vid := msg.(MessageIncEvalUpdate).vid
	//			dist := msg.(MessageIncEvalUpdate).dist
	//			log.Info().Msgf("Receive update message: from=%d, round=%d, vid=%d, dist=%f",
	//				from, round, vid, dist)
	//			if min > round {
	//				min = round
	//			}
	//			if max < round {
	//				max = round
	//			}
	//			if _, exists := updateMap[vid]; exists {
	//				if dist < updateMap[vid] {
	//					updateMap[vid] = dist
	//				}
	//			} else {
	//				updateMap[vid] = dist
	//			}
	//		}
	//		log.Info().Msgf("max: %d, min: %d, now: %d", max, min, w.round)
	//		updateMap = path.IncEvalDijkstraFrom(updateMap, &w.shortest, w.g.Vertex(0), w.g)
	//		log.Info().Msgf("IncEval #%d: %+v, update=%+v", w.round, w.shortest, updateMap)
	//		for vid, dist := range updateMap {
	//			peer := w.peers[w.fot[w.g.Vertex(vid)]]
	//			if err := w.SendMessage(peer, MessageIncEvalUpdate{
	//				from:  mid,
	//				round: w.round,
	//				vid:   vid,
	//				dist:  dist,
	//			}); err != nil {
	//				panic(err)
	//			}
	//		}
	//		// TODO: adjust DS
	//		w.round += 1
	//		if len(w.Receive(opcodeIncEvalUpdate)) == 0 {
	//			w.isInactive = 1
	//			goto WaitTerminateRequest
	//		}
	//	}
	//}

//WaitTerminateRequest:
	msg = <-w.Receive(opcodeTerminateRequest)
	log.Info().Msgf("Receive Terminate Request from master %d", msg.(MessageTerminateRequest).from)
	if len(w.Receive(opcodeIncEvalUpdate)) != 0 {
		w.isInactive = 0
		if err := w.SendMessage(w.master, MessageTerminateNACK{from: w.id}); err != nil {
			panic(err)
		}
		goto IncrementalEvaluation
	} else {
		if err := w.SendMessage(w.master, MessageTerminateACK{from: w.id}); err != nil {
			panic(err)
		}
	}

	msg = <-w.Receive(opcodeAssembleRequest)
	log.Info().Msgf("Receive Assemble Request from master %d", msg.(MessageAssembleRequest).from)
	if len(w.Receive(opcodeIncEvalUpdate)) == 0 {
		if err := w.SendMessage(w.master, MessageAssembleResponse{
			from:   w.id,
			result: w.shortest.WeightToAllVertices(),
		}); err != nil {
			panic(err)
		}
	} else {
		goto IncrementalEvaluation
	}

	log.Info().Msg("Done!")
}

func (w *Worker) ID() uint64 {
	return w.id
}

func (w *Worker) Host() string {
	return w.host
}

func (w *Worker) Port() uint16 {
	return w.port
}

func (w *Worker) Address() string {
	return fmt.Sprintf("%s:%d", w.host, w.port)
}

func (w *Worker) Peers() map[uint64]*Peer {
	return w.peers
}

func (w *Worker) Round() uint64 {
	return w.round
}

func (w *Worker) Master() *Peer {
	return w.master
}

func (w *Worker) Graph() graph.Graph {
	return w.g
}

func (w *Worker) Shortest() path.Shortest {
	return w.shortest
}

func (w *Worker) IsInactive() uint32 {
	return atomic.LoadUint32(&w.isInactive)
}
