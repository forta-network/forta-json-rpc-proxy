package clients

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

type builderClient struct {
	rpcClient *rpc.Client
}

// NewBuilderClient creates a new bundler client which sends bundles to a builder.
func NewBuilderClient(ctx context.Context, rawUrl string) (*builderClient, error) {
	c, err := rpc.DialContext(ctx, rawUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to dial builder rpc: %v", err)
	}
	return &builderClient{rpcClient: c}, nil
}

// SendBundle sends a bundle of transactions to a builder.
func (bc *builderClient) SendBundle(ctx context.Context, txs []hexutil.Bytes) error {
	return bc.rpcClient.CallContext(ctx, nil, "eth_sendBundle", struct {
		Txs []hexutil.Bytes `json:"txs"`
	}{
		Txs: txs,
	})
}
