package service

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

const contractErrAttestationNotFound = "120a2e773951f3178d454b2ed5973f0df81a0d0dc37028cedef36e011a64a265"

type wrapperService struct {
	chainID        *big.Int
	rpcClient      RPCClient
	ethClient      EthClient
	bundler        Bundler
	attester       Attester
	enableBundling bool
}

// NewWrapperService creates a new service that wraps a few JSON-RPC methods.
func NewWrapperService(
	chainID *big.Int, rpcClient RPCClient, ethClient EthClient,
	bundler Bundler, attester Attester,
) *wrapperService {
	return &wrapperService{
		chainID:   chainID,
		rpcClient: rpcClient,
		ethClient: ethClient,
		bundler:   bundler,
		attester:  attester,
	}
}

// Frontrunning:

func (s *wrapperService) SendRawTransaction(ctx context.Context, userTx hexutil.Bytes) (common.Hash, error) {
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
	_, err = s.Call(ctx, txToArgs(signer, tx), nil, nil, nil)
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

func (s *wrapperService) sendTx(ctx context.Context, tx hexutil.Bytes) (common.Hash, error) {
	return s.ethClient.SendRawTransaction(ctx, tx)
}

func txToArgs(signer common.Address, tx *types.Transaction) (txArgs TransactionArgs) {
	txArgs.From = &signer
	txArgs.To = tx.To()
	gas := tx.Gas()
	txArgs.Gas = (*hexutil.Uint64)(&gas)
	txArgs.GasPrice = (*hexutil.Big)(tx.GasPrice())
	txArgs.Value = (*hexutil.Big)(tx.Value())
	data := tx.Data()
	txArgs.Data = (*hexutil.Bytes)(&data)
	return
}

// State overridden calls:

func (s *wrapperService) Call(ctx context.Context, txArgs TransactionArgs, blockNrOrHash *rpc.BlockNumberOrHash, stateOverride *StateOverride, blockOverrides *BlockOverrides) (result hexutil.Bytes, err error) {
	stateOverride = AddFortaFirewallStateOverride(stateOverride)
	err = s.rpcClient.CallContext(ctx, &result, "eth_call", txArgs, blockNrOrHash, stateOverride, blockOverrides)
	return
}

func (s *wrapperService) EstimateGas(ctx context.Context, txArgs TransactionArgs, blockNrOrHash *rpc.BlockNumberOrHash, stateOverride *StateOverride) (result interface{}, err error) {
	stateOverride = AddFortaFirewallStateOverride(stateOverride)
	err = s.rpcClient.CallContext(ctx, &result, "eth_estimateGas", txArgs, blockNrOrHash, stateOverride)
	return
}
