package aap_v2

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/zakimal/aap.v2/log"
	"reflect"
	"sync"
)

type Opcode byte // 8 bit

const OpcodeNil Opcode = 0

var (
	// Opcode => Message
	opcodes map[Opcode]Message

	// Message => Opcode
	messages map[reflect.Type]Opcode

	mutex sync.Mutex

	// AAP opcodes
	opcodeHelloWorker      Opcode
	opcodeHelloMaster      Opcode
	opcodePEvalRequest     Opcode
	opcodePEvalResponse    Opcode
	opcodeIncEvalUpdate    Opcode
	opcodeNotifyInactive   Opcode
	opcodeTerminateRequest Opcode
	opcodeTerminateACK     Opcode
	opcodeTerminateNACK    Opcode
	opcodeAssembleRequest  Opcode
	opcodeAssembleResponse Opcode
)

// Byte() returns byte representation of opcode.
func (op Opcode) Byte() [1]byte {
	var b [1]byte
	b[0] = byte(op)
	return b
}

// NextAvailableOpcode() returns next free opcode.
func NextAvailableOpcode() Opcode {
	mutex.Lock()
	defer mutex.Unlock()
	return Opcode(len(opcodes))
}

// for debug
// DebugOpcodes() prints a list of registered opcode: message pair so far.
func DebugOpcodes() {
	mutex.Lock()
	defer mutex.Unlock()
	log.Debug().Msg("Here are all opcodes registered so far")
	for op, msg := range opcodes {
		fmt.Printf("\t[%d] <==> %s\n", op, reflect.TypeOf(msg).String())
	}
}

// MessageFromOpcode() returns Message interface corresponding to given Opcode.
func MessageFromOpcode(op Opcode) (Message, error) {
	mutex.Lock()
	defer mutex.Unlock()
	typ, exist := opcodes[op]
	if !exist {
		return nil, errors.Errorf("There is no messageReceiver type registered to opcode [%d]\n", op)
	}
	msg, ok := reflect.New(reflect.TypeOf(typ)).Elem().Interface().(Message)
	if !ok {
		return nil, errors.Errorf("Invalid messageReceiver type associated to opcode [%d]\n", op)
	}
	return msg, nil
}

// OpcodeFromMessage() returns Opcode corresponding to given Message interface.
func OpcodeFromMessage(msg Message) (Opcode, error) {
	mutex.Lock()
	defer mutex.Unlock()
	typ := reflect.TypeOf(msg)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	op, exist := messages[typ]
	if !exist {
		return OpcodeNil, errors.Errorf("There is no opcode registered for messageReceiver type %v\n", typ)
	}
	return op, nil
}

// RegisterMessage() registers given Opcode and Message.
func RegisterMessage(op Opcode, msg interface{}) Opcode {
	typ := reflect.TypeOf(msg).Elem()
	mutex.Lock()
	defer mutex.Unlock()
	if opcode, registered := messages[typ]; registered {
		return opcode
	}
	opcodes[op] = reflect.New(typ).Elem().Interface().(Message)
	messages[typ] = op
	return op
}

func resetOpcodes() {
	mutex.Lock()
	defer mutex.Unlock()
	opcodes = map[Opcode]Message{
		OpcodeNil: reflect.New(reflect.TypeOf((*MessageNil)(nil)).Elem()).Elem().Interface().(Message),
	}
	messages = map[reflect.Type]Opcode{
		reflect.TypeOf((*MessageNil)(nil)).Elem(): OpcodeNil,
	}
}
