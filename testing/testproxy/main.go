package main

import (
	"github.com/forta-network/forta-json-rpc-proxy/proxy"
	"github.com/forta-network/forta-json-rpc-proxy/service"
	"github.com/forta-network/forta-json-rpc-proxy/testing/fake"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var cfg struct {
	service.Config
	ValidatorAddress   string `envconfig:"VALIDATOR_ADDRESS"`
	AttesterPrivateKey string `envconfig:"ATTESTER_PRIVATE_KEY"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.WithError(err).Info("failed to load .env file - safely continuing with environment defaults")
	}

	err = envconfig.Process("forta-json-rpc-proxy", &cfg)
	if err != nil {
		logrus.WithError(err).Panic("failed to read config")
	}

	proxy.StartWithAttester(cfg.Config, fake.NewAttester(cfg.TargetRPCURL, cfg.ValidatorAddress, cfg.AttesterPrivateKey))
}
