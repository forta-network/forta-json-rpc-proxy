package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

const contractErrAttestationNotFound = "120a2e773951f3178d454b2ed5973f0df81a0d0dc37028cedef36e011a64a265"

var fortaFirewallBypassFlag = json.RawMessage([]byte(`{"0x0000000000000000000000000000000000f01274":{"code": "0x10"}}`))

type Service struct {
	chainID        *big.Int
	rpcClient      RPCClient
	ethClient      EthClient
	bundler        Bundler
	attester       Attester
	enableBundling bool
}

// NewService creates a new service.
func NewService(
	chainID *big.Int, rpcClient RPCClient, ethClient EthClient,
	bundler Bundler, attester Attester,
) *Service {
	return &Service{
		chainID:   chainID,
		rpcClient: rpcClient,
		ethClient: ethClient,
		bundler:   bundler,
		attester:  attester,
	}
}

// The complete list for reference:
// eth_sendRawTransaction

// eth_call
// eth_estimateGas

// eth_getBalance
// eth_getTransactionCount
// eth_getBlockByNumber
// eth_getBlockByHash
// eth_getBlockNumber
// eth_getCode
// eth_gasPrice
// eth_getTransactionReceipt

// Frontrunning:

func (s *Service) SendRawTransaction(ctx context.Context, userTx hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(userTx); err != nil {
		return common.Hash{}, err
	}
	signer, err := types.LatestSignerForChainID(s.chainID).Sender(tx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to recover tx signer: %v", err)
	}

	// Refuse to attest to safety of transactions that deploy contracts.
	if tx.To() == nil {
		return s.sendTx(ctx, userTx)
	}
	// See if the the tx requires an attestation.
	_, err = s.Call(ctx, ethereum.CallMsg{
		From:       signer,
		To:         tx.To(),
		Gas:        tx.Gas(),
		GasPrice:   tx.GasPrice(),
		GasFeeCap:  tx.GasFeeCap(),
		GasTipCap:  tx.GasTipCap(),
		Value:      tx.Value(),
		Data:       tx.Data(),
		AccessList: tx.AccessList(),
	}, nil)
	// If it doesn't need attestation - just send the tx
	if err == nil || !strings.Contains(err.Error(), contractErrAttestationNotFound) {
		return s.sendTx(ctx, userTx)
	}

	// The attester should give back a transaction.
	attestTx, err := s.attester.AttestWithTx(ctx, &AttestRequest{
		From:    signer,
		To:      *tx.To(),
		Input:   hexutil.Bytes(tx.Data()).String(),
		Value:   (*hexutil.Big)(tx.Value()),
		ChainID: s.chainID.Uint64(),
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("attestation fails: %v", err)
	}

	// Send both txs in a bundle.
	if err := s.bundler.SendBundle(ctx, []hexutil.Bytes{attestTx, userTx}); err != nil {
		return common.Hash{}, fmt.Errorf("failed to send bundle: %v", err)
	}
	return tx.Hash(), nil
}

func (s *Service) sendTx(ctx context.Context, tx hexutil.Bytes) (common.Hash, error) {
	return s.ethClient.SendRawTransaction(ctx, tx)
}

// State overridden calls:

func (s *Service) Call(ctx context.Context, args ...interface{}) (result hexutil.Bytes, err error) {
	err = s.rpcClient.CallContext(ctx, &result, "eth_call", append(args, fortaFirewallBypassFlag)...)
	return
}

func (s *Service) EstimateGas(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	err = s.rpcClient.CallContext(ctx, &result, "eth_estimateGas", append(args, fortaFirewallBypassFlag)...)
	return
}

// Whitelisted and defaulted calls:

func (s *Service) callCtx(ctx context.Context, method string, args ...interface{}) (result interface{}, err error) {
	err = s.rpcClient.CallContext(ctx, &result, method, args...)
	return
}

func (s *Service) GetBalance(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	return s.callCtx(ctx, "eth_getBalance", args...)
}

func (s *Service) GetTransactionCount(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	return s.callCtx(ctx, "eth_getTransactionCount", args...)
}

func (s *Service) GetBlockByNumber(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	return s.callCtx(ctx, "eth_getBlockByNumber", args...)
}

func (s *Service) GetBlockByHash(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	return s.callCtx(ctx, "eth_getBlockByHash", args...)
}

func (s *Service) GetBlockNumber(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	return s.callCtx(ctx, "eth_getBlockNumber", args...)
}

func (s *Service) GetCode(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	return s.callCtx(ctx, "eth_getCode", args...)
}

func (s *Service) GasPrice(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	return s.callCtx(ctx, "eth_gasPrice", args...)
}

func (s *Service) GetTransactionReceipt(ctx context.Context, args ...interface{}) (result interface{}, err error) {
	return s.callCtx(ctx, "eth_getTransactionReceipt", args...)
}
