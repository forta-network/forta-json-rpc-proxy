package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/forta-network/forta-json-rpc-proxy/clients"
	"github.com/forta-network/forta-json-rpc-proxy/service"
	"github.com/forta-network/forta-json-rpc-proxy/utils"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, _ := utils.InitMainContext()

	err := godotenv.Load()
	if err != nil {
		logrus.WithError(err).Info("failed to load .env file - continuing to run with default env vars")
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})

	var cfg service.Config
	err = envconfig.Process("forta-json-rpc-proxy", &cfg)
	if err != nil {
		logrus.WithError(err).Panic("failed to read config")
	}

	logrus.SetLevel(cfg.LogLevel)

	ethClient, err := ethclient.DialContext(ctx, cfg.TargetRPCURL)
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

	srv := service.NewWrapperService(chainID, rpcClient, wrappedClient, bundler, attesterClient)
	if err != nil {
		logrus.WithError(err).Panic("failed to create service")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin", "Authorization"},
		ExposedHeaders:   []string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin", "Authorization"},
		AllowCredentials: true,
	})

	err = utils.ListenAndServe(ctx, &http.Server{
		Handler:      c.Handler(service.NewProxy(srv, cfg.TargetRPCURL, cfg.APIKey)),
		Addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}, fmt.Sprintf("started forta json-rpc proxy for chain %d", chainID.Uint64()))
	if err != nil {
		logrus.WithError(err).Error("http server returned error")
	}
}
