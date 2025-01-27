package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/forta-network/forta-json-rpc-proxy/clients"
	"github.com/forta-network/forta-json-rpc-proxy/interfaces"
	"github.com/forta-network/forta-json-rpc-proxy/service"
	"github.com/forta-network/forta-json-rpc-proxy/utils"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

// Start is a blocking function which initializes internal dependencies, services
// and the proxy and listens for incoming requests.
func Start(cfg service.Config) {
	attesterClient := clients.NewAttesterClient(cfg.AttesterAPIURL, cfg.AttesterAuthToken)
	StartWithAttester(cfg, attesterClient)
}

// StartWithAttester starts with given attester implementation.
func StartWithAttester(cfg service.Config, attester interfaces.Attester) {
	ctx, _ := utils.InitMainContext()
	logrus.SetFormatter(&logrus.JSONFormatter{})
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

	var bundler interfaces.Bundler
	if len(cfg.BuilderAPIURL) > 0 {
		bundler, err = clients.NewBuilderClient(ctx, cfg.BuilderAPIURL)
		if err != nil {
			logrus.WithError(err).Panic("failed to create new builder client")
		}
	} else {
		bundler = clients.NewTxSender(wrappedClient, cfg.TxRetryTimes, cfg.TxRetryIntervalSeconds)
	}

	srv := service.NewWrapperService(chainID, rpcClient, wrappedClient, bundler, attester)
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
