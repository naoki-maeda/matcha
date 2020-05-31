package blockchain

import (
	"context"

	"github.com/btcsuite/btcd/rpcclient"
	zmq "github.com/go-zeromq/zmq4"
)

// ZMQ topics
const (
	Hash      = "hash"
	HashBlock = "hashblock"
	HashTx    = "hashtx"
	RawBlock  = "rawblock"
	RawTx     = "rawtx"
)

var topics = []string{Hash, HashBlock, HashTx, RawBlock, RawTx}

// ZMQ zero mq settings
// https://github.com/bitcoin/bitcoin/blob/master/doc/zmq.md
type ZMQ struct {
	client zmq.Socket
}

// Message is ZMQ message
type Message struct {
	Msg zmq.Msg
}

// NewZmqClient zero mq client
func NewZmqClient(zmqAddress string, verbose bool) (*ZMQ, error) {
	client := zmq.NewSub(context.Background())

	err := client.Dial(zmqAddress)
	if err != nil {
		return nil, err
	}
	// set topics
	for _, topic := range topics {
		if err := client.SetOption(zmq.OptionSubscribe, topic); err != nil {
			return nil, err
		}
	}
	zmq := &ZMQ{client}
	return zmq, nil
}

// Close socket connection close
func (zmq *ZMQ) Close() {
	zmq.client.Close()
}

// Receive zmq message
func (zmq *ZMQ) Receive(message chan Message) error {
	defer zmq.Close()
	for {
		msg, err := zmq.client.Recv()
		if err != nil {
			return err
		}
		message <- Message{msg}
	}
}
