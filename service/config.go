package service

import (
	"github.com/sirupsen/logrus"
)

// Config is the service config.
type Config struct {
	LogLevel               logrus.Level `default:"info" envconfig:"LOG_LEVEL"`
	Port                   int          `default:"8080" envconfig:"PORT"`
	Target                 string       `required:"true" envconfig:"TARGET"`
	AttesterAPIURL         string       `required:"true" envconfig:"ATTESTER_API_URL"`
	AttesterAuthToken      string       `required:"true" envconfig:"ATTESTER_AUTH_TOKEN"`
	BuilderAPIURL          string       `envconfig:"BUILDER_API_URL"`
	TxRetryTimes           int          `default:"10" envconfig:"TX_RETRY_TIMES"`
	TxRetryIntervalSeconds int          `default:"2" envconfig:"TX_RETRY_INTERVAL_SECONDS"`
}
