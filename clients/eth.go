package clients

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ethClient struct {
	*ethclient.Client
}

// NewEthClient is an Ethereum client wrapper.
func NewEthClient(ec *ethclient.Client) *ethClient {
	return &ethClient{Client: ec}
}

// SendRawTransaction sends the transaction and returns the hash.
func (ec *ethClient) SendRawTransaction(ctx context.Context, tx hexutil.Bytes) (h common.Hash, err error) {
	err = ec.Client.Client().CallContext(ctx, &h, "eth_sendRawTransaction", tx)
	return
}
