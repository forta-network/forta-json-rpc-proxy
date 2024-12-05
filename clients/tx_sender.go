package clients

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/forta-network/forta-json-rpc-proxy/service"
	"github.com/sirupsen/logrus"
)

type txSender struct {
	ethClient     service.EthClient
	retryTimes    int
	retryInterval time.Duration
}

var _ service.Bundler = &txSender{}

// NewTxSender creates a new bundler client which sends transactions in order.
func NewTxSender(ethClient service.EthClient, retryTimes int, retryIntervalSeconds int) *txSender {
	return &txSender{
		ethClient:     ethClient,
		retryTimes:    retryTimes,
		retryInterval: time.Duration(retryIntervalSeconds) * time.Second,
	}
}

// SendBundle sends a bundle of transactions in correct order, one after another.
// Currently it is implemented to support only two transactions.
func (ts *txSender) SendBundle(ctx context.Context, txs []hexutil.Bytes) error {
	if len(txs) != 2 {
		return errors.New("unexpected bundle size")
	}
	txHash, err := ts.ethClient.SendRawTransaction(ctx, txs[0])
	if err != nil {
		return fmt.Errorf("failed to send first tx: %v", err)
	}

	// Wait for just a second.
	time.Sleep(time.Second)

	// Try to get the receipt of the first tx.
	for i := 0; i < ts.retryTimes; i++ {
		receipt, err := ts.ethClient.TransactionReceipt(ctx, txHash)
		if err != nil {
			logrus.WithError(err).Debug("failed to get first tx receipt")
			time.Sleep(ts.retryInterval)
			continue
		}
		if receipt.Status != 1 {
			return fmt.Errorf("first tx failed: %s", txHash.Hex())
		}
		break
	}

	// Send the second transaction and forget about it.
	_, err = ts.ethClient.SendRawTransaction(ctx, txs[1])
	return err
}
