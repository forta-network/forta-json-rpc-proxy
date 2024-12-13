package main

import (
	"github.com/forta-network/forta-json-rpc-proxy/proxy"
	"github.com/forta-network/forta-json-rpc-proxy/service"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.WithError(err).Info("failed to load .env file - safely continuing with environment defaults")
	}

	var cfg service.Config
	err = envconfig.Process("forta-json-rpc-proxy", &cfg)
	if err != nil {
		logrus.WithError(err).Panic("failed to read config")
	}

	proxy.Start(cfg)
}
