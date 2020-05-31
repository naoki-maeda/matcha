package blockchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/rpcclient"
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

var topics = []string{hash, hashBlock, hashTx, rawBlock, rawTx}

// ZMQ zero mq settings
// https://github.com/bitcoin/bitcoin/blob/master/doc/zmq.md
type ZMQ struct {
	socket    zmq.Socket
	isRunning bool
	finished  chan error
	verbose   bool
}

// NewZmqClient zero mq client
func NewZmqClient(zmqAddress string, verbose bool) (*ZMQ, error) {
	socket := zmq.NewSub(context.Background())

	err := socket.Dial(zmqAddress)
	if err != nil {
		return nil, err
	}
	// set topics
	for _, topic := range topics {
		if err := socket.SetOption(zmq.OptionSubscribe, topic); err != nil {
			return nil, err
		}
	}
	zmq := &ZMQ{socket, true, make(chan error), verbose}
	return zmq, nil
}

// Close socket connection close
func (zmq *ZMQ) Close() {
	zmq.socket.Close()
}

// Sync zmq Syncing
func (zmq *ZMQ) Sync(rpc *rpcclient.Client) error {
	defer zmq.Close()
	for {
		msg, err := zmq.socket.Recv()
		if err != nil {
			return err
		}
		// maybe msg frames are fixed 3 length
		if msg.Frames == nil || len(msg.Frames) != 3 {
			return errors.New("ZMQ receive message is nil or a 3 length")
		}
		topic := fmt.Sprintf("%s", msg.Frames[0])
		switch topic {
		case rawBlock:
			block, err := btcutil.NewBlockFromBytes(msg.Frames[1])
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("==================")
			fmt.Printf("block hash: %s\n", block.Hash())
			if zmq.verbose == true {
				blockVerbose, err := rpc.GetBlockVerbose(block.Hash())
				if err != nil {
					return err
				}
				j, err := json.Marshal(blockVerbose)
				if err != nil {
					return err
				}
				fmt.Println(string(j))
			}
		case rawTx:
			tx, err := btcutil.NewTxFromBytes(msg.Frames[1])
			if err != nil {
				return err
			}
			fmt.Println("==================")
			fmt.Printf("tx hash: %s\n", tx.Hash())
			if zmq.verbose == true {
				txVerbose, err := rpc.GetRawTransactionVerbose(tx.Hash())
				if err != nil {
					return err
				}
				j, err := json.Marshal(txVerbose)
				if err != nil {
					return err
				}
				fmt.Println(string(j))
			}
		}
	}
}
