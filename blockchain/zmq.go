package blockchain

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcutil"
	zmq "github.com/go-zeromq/zmq4"
)

const (
	hash      = "hash"
	hashBlock = "hashblock"
	hashTx    = "hashtx"
	rawBlock  = "rawblock"
	rawTx     = "rawtx"
)

// ZMQ zero mq settings
// https://github.com/bitcoin/bitcoin/blob/master/doc/zmq.md
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
	if err := socket.SetOption(zmq.OptionSubscribe, hash); err != nil {
		return nil, err
	}
	if err := socket.SetOption(zmq.OptionSubscribe, hashBlock); err != nil {
		return nil, err
	}
	if err := socket.SetOption(zmq.OptionSubscribe, hashTx); err != nil {
		return nil, err
	}
	if err := socket.SetOption(zmq.OptionSubscribe, rawBlock); err != nil {
		return nil, err
	}
	if err := socket.SetOption(zmq.OptionSubscribe, rawTx); err != nil {
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
		topic := fmt.Sprintf("%s", msg.Frames[0])
		switch topic {
		case rawBlock:
			block, err := btcutil.NewBlockFromBytes(msg.Frames[1])
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println(`Generate Block
==================`)
			// height is -1 ?
			fmt.Printf("height: %d hash: %s\n", block.Height(), block.Hash())
		case rawTx:
			tx, err := btcutil.NewTxFromBytes(msg.Frames[1])
			if err != nil {
				return err
			}
			fmt.Println(`Transaction
==================`)
			fmt.Printf("hash: %s\n", tx.Hash())
		}
	}
}
