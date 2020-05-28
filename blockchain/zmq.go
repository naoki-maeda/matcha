package blockchain

import (
	"fmt"
	"context"
	zmq "github.com/go-zeromq/zmq4"
)

const (
	hash = "hash"
	hashBlock = "hashblock"
	hashTx = "hashtx"
)


// ZMQ zero mq settings
type ZMQ struct {
	socket    zmq.Socket
	isRunning bool
	finished  chan error
}

// NewZmqClient zero mq client
func NewZmqClient(zmqAddress string) (*ZMQ, error) {
	socket := zmq.NewSub(context.Background())

	err := socket.Dial(zmqAddress)
	if err != nil {
		return nil, err
	}
	err = socket.SetOption(zmq.OptionSubscribe, hash)
	if err != nil {
		return nil, err
	}
	zmq := &ZMQ{socket, true, make(chan error)}
	return zmq, nil
}

// Close socket connection close
func (zmq *ZMQ) Close() {
	zmq.socket.Close()
}

// Sync zmq Syncing
func (zmq *ZMQ) Sync() error {
	defer zmq.Close()
	for {
		msg, err := zmq.socket.Recv()
		if err != nil {
			return err
		}
		fmt.Println(msg.String())
	}
}
