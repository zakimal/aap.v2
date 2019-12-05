package aap_v2

import (
	"github.com/pkg/errors"
	"github.com/zakimal/aap.v2/payload"
)

type Message interface {
	Read(reader payload.Reader) (Message, error)
	Write() []byte
}

type MessageNil struct{}

func (MessageNil) Read(reader payload.Reader) (Message, error) {
	return MessageNil{}, nil
}
func (m MessageNil) Write() []byte {
	return nil
}

type MessageHelloWorker struct {
	from uint64
}

func (MessageHelloWorker) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageHelloWorker`")
	}
	return MessageHelloWorker{from: from}, nil
}
func (m MessageHelloWorker) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

type MessageHelloMaster struct {
	from uint64
}

func (MessageHelloMaster) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageHelloMaster`")
	}
	return MessageHelloMaster{from: from}, nil
}
func (m MessageHelloMaster) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

type MessagePEvalRequest struct {
	from uint64
}

func (MessagePEvalRequest) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessagePEvalRequest`")
	}
	return MessagePEvalRequest{from: from}, nil
}
func (m MessagePEvalRequest) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

type MessagePEvalResponse struct {
	from uint64
}

func (MessagePEvalResponse) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessagePEvalResponse`")
	}
	return MessagePEvalResponse{from: from}, nil
}
func (m MessagePEvalResponse) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

// TODO: dist?
type MessageIncEvalUpdate struct {
	from  uint64
	round uint64
	vid   int64
	dist  float64
}

func (MessageIncEvalUpdate) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageIncEvalUpdate`")
	}
	round, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `round` of `MessageIncEvalUpdate`")
	}
	nid, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `vid` of `MessageIncEvalUpdate`")
	}
	dist, err := reader.ReadFloat64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `dist` of `MessageIncEvalUpdate`")
	}
	return MessageIncEvalUpdate{
		from:  from,
		round: round,
		vid:   int64(nid),
		dist:  dist,
	}, nil
}
func (m MessageIncEvalUpdate) Write() []byte {
	return payload.NewWriter(nil).
		WriteUint64(m.from).
		WriteUint64(m.round).
		WriteUint64(uint64(m.vid)).
		WriteFloat64(m.dist).
		Bytes()
}

type MessageNotifyInactive struct {
	from uint64
}

func (MessageNotifyInactive) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageNotifyInactive`")
	}
	return MessageNotifyInactive{from: from}, nil
}
func (m MessageNotifyInactive) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

type MessageTerminateRequest struct {
	from uint64
}

func (MessageTerminateRequest) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageTerminateRequest`")
	}
	return MessageTerminateRequest{from: from}, nil
}
func (m MessageTerminateRequest) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

type MessageTerminateACK struct {
	from uint64
}

func (MessageTerminateACK) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageTerminateACK`")
	}
	return MessageTerminateACK{from: from}, nil
}
func (m MessageTerminateACK) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

type MessageTerminateNACK struct {
	from uint64
}

func (MessageTerminateNACK) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageTerminateACK`")
	}
	return MessageTerminateACK{from: from}, nil
}
func (m MessageTerminateNACK) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

type MessageAssembleRequest struct {
	from uint64
}

func (MessageAssembleRequest) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageAssembleRequest`")
	}
	return MessageAssembleRequest{from: from}, nil
}
func (m MessageAssembleRequest) Write() []byte {
	return payload.NewWriter(nil).WriteUint64(m.from).Bytes()
}

type MessageAssembleResponse struct {
	from   uint64
	result map[int64]float64
}

func (m MessageAssembleResponse) Read(reader payload.Reader) (Message, error) {
	from, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `from` of `MessageAssembleResponse`")
	}
	length, err := reader.ReadUint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read `len(result)` of `MessageAssembleResponse`")
	}
	result := make(map[int64]float64)
	var i uint64
	for i = 0; i < length; i++ {
		nid, err := reader.ReadUint64()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read %d th key of `result` of `MessageAssembleResponse`", i)
		}
		data, err := reader.ReadFloat64()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read %d th value of `result` of `MessageAssembleResponse`", i)
		}
		result[int64(nid)] = data
	}
	return MessageAssembleResponse{
		from:   from,
		result: result,
	}, nil
}
func (m MessageAssembleResponse) Write() []byte {
	writer := payload.NewWriter(nil)
	writer.WriteUint64(m.from)
	writer.WriteUint64(uint64(len(m.result)))
	for nid, data := range m.result {
		writer.WriteUint64(uint64(nid))
		writer.WriteFloat64(data)
	}
	return writer.Bytes()
}
