package nonce

import (
	"context"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/forta-network/forta-json-rpc-proxy/interfaces"
)

type nonceClient struct {
	interfaces.EthClient
	managedAcc string
	maxDrift   uint64

	localNonce uint64
	mu         sync.Mutex
}

// NewNonceManagerClient manages the nonce of a specific account by comparing with remote and
// incrementing locally.
// TODO: The local nonce can be preserved at an external persistent cache (e.g. Redis) in order to
// prevent unwanted nonce resets when the service running this code is restarted.
func NewNonceManagerClient(ethClient interfaces.EthClient, managedAcc string, maxDrift uint64) *nonceClient {
	return &nonceClient{
		EthClient:  ethClient,
		managedAcc: managedAcc,
		maxDrift:   maxDrift,
	}
}

func (nc *nonceClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	remoteNonce, err := nc.EthClient.NonceAt(ctx, account, blockNumber)
	if err != nil {
		return 0, err
	}
	// Fall back to remote nonce if requested account is not the one that is managed by this client.
	if !strings.EqualFold(nc.managedAcc, account.Hex()) {
		return remoteNonce, err
	}

	nc.mu.Lock()
	defer nc.mu.Unlock()

	// Set the nonce for the first time.
	if nc.localNonce == 0 {
		nc.localNonce = remoteNonce
	}
	// Fall back to the highest nonce.
	if remoteNonce > nc.localNonce {
		nc.localNonce = remoteNonce
	}
	// Reset nonce to remote value when local nonce reaches to max drift.
	if nc.localNonce-remoteNonce >= nc.maxDrift {
		nc.localNonce = remoteNonce
	}
	// Consider that this nonce will be used for a to-be-sent transaction and consume it.
	// This can be error prone if the nonce is not consumed by the requester and that is tolerated
	// by the nonce reset logic.
	currNonce := nc.localNonce
	nc.localNonce++
	return currNonce, nil
}
