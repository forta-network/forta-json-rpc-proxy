package interfaces

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type RPCClient interface {
	Close()
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
}

type EthClient interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	SendRawTransaction(ctx context.Context, tx hexutil.Bytes) (common.Hash, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
}

type Bundler interface {
	SendBundle(ctx context.Context, txs []hexutil.Bytes) error
}

type AttestRequest struct {
	From    common.Address `json:"from"`
	To      common.Address `json:"to"`
	Input   string         `json:"input"`
	Value   *hexutil.Big   `json:"value"`
	ChainID uint64         `json:"chainId"`
}

type AttesterError error

var (
	ErrAttestationNotRequired AttesterError = errors.New("attestation not required")
)

type Attester interface {
	AttestWithTx(ctx context.Context, req *AttestRequest) (tx hexutil.Bytes, err error)
}
