package blockchain

import (
	"github.com/btcsuite/btcd/rpcclient"
)

// MaxTries is RPC max tries count
var MaxTries int64 = 10

// RPC client & settings
type RPC struct {
	Client *rpcclient.Client
}

// NewRPC return new RPC client
func NewRPC(hostName, port, user, password string, disabletls bool) (*RPC, error) {
	host := hostName + ":" + port
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         password,
		HTTPPostMode: true,       // Bitcoin core only supports HTTP POST mode
		DisableTLS:   disabletls, // Bitcoin core does not provide TLS by default
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}
	return &RPC{client}, err
}
