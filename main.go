package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/forta-network/forta-json-rpc-proxy/clients"
	"github.com/forta-network/forta-json-rpc-proxy/service"
	"github.com/forta-network/forta-json-rpc-proxy/utils"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, _ := utils.InitMainContext()

	logrus.SetFormatter(&logrus.JSONFormatter{})

	var cfg service.Config
	err := envconfig.Process("forta-json-rpc-proxy", &cfg)
	if err != nil {
		logrus.WithError(err).Panic("failed to read config")
	}

	logrus.SetLevel(cfg.LogLevel)

	ethClient, err := ethclient.DialContext(ctx, cfg.Target)
	if err != nil {
		logrus.WithError(err).Panic("failed to dial target rpc")
	}
	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		logrus.WithError(err).Panic("failed to get chain id")
	}
	rpcClient := ethClient.Client()

	wrappedClient := clients.NewEthClient(ethClient)

	var bundler service.Bundler
	if len(cfg.BuilderAPIURL) > 0 {
		bundler, err = clients.NewBuilderClient(ctx, cfg.BuilderAPIURL)
		if err != nil {
			logrus.WithError(err).Panic("failed to create new builder client")
		}
	} else {
		bundler = clients.NewTxSender(wrappedClient, cfg.TxRetryTimes, cfg.TxRetryIntervalSeconds)
	}

	attesterClient := clients.NewAttesterClient(cfg.AttesterAPIURL, cfg.AttesterAuthToken)

	srv := service.NewService(chainID, rpcClient, wrappedClient, bundler, attesterClient)
	if err != nil {
		logrus.WithError(err).Panic("failed to create service")
	}

	err = utils.ListenAndServe(ctx, &http.Server{
		Handler:      service.NewProxy(srv),
		Addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}, fmt.Sprintf("started forta json-rpc proxy for chain %d", chainID.Uint64()))
	if err != nil {
		logrus.WithError(err).Error("http server returned error")
	}
}
