/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Config struct {
	LogLevel                string        `env:"LOG" envDefault:"info"`
	PollInterval            string        `env:"POLL_INTERVAL" envDefault:"1000ms"`
	PingTimeout             string        `env:"POLL_INTERVAL" envDefault:"1000ms"`
	RedisPass               string        `env:"REDIS_PASS" envDefault:""`
	RedisAddrs              string        `env:"REDIS_ADDRS" envDefault:"0.0.0.0:6379"`
	RedisTLS                bool          `env:"REDIS_TLS" envDefault:"false"`                                      // enable TLS for communication with Redis
	RedisRootCACerts        string        `env:"REDIS_ROOTCA_CERTS" envDefault:""`                                  // comma-separated root CA's certificates list for TLS with Redis
	ServiceStorageNamespace string        `env:"SERVICES_NAMESPACES" envDefault:"sys-robots-list,sys-parsers-list"` // keys in storage where services lists are stored
	DistributionNamespace   string        `env:"DISTRIBUTION_NAMESPACE" envDefault:"sys-matching-table"`            // key in storage where work distribution data is stored
	WorkunitsNamespace      string        `env:"WORKUNITS_NAMESPACE" envDefault:"sys-channels"`                     // key in storage where work units list is stored
	PromPort                uint          `env:"PROM_PORT" envDefault:"9090"`                                       // Prometheus metrics port
	KATime                  time.Duration `env:"KA_TIME"`                                                           // KeepAlive time
	KATimeout               time.Duration `env:"KA_TIMEOUT"`                                                        // KeepAlive timeout
	KAPermitWithoutStream   bool          `env:"KA_PERMIT_WITHOUT_STREAM" envDefault:"false"`                       // KeepAlive param: if true, client sends keepalive pings even with no active RPCs; if false, when there are no active RPCs, Time and Timeout will be ignored and no keepalive pings will be sent
	typeOfConfig            string
}

func GetConfig() (*Config, error) {
	var err error

	conf := cliArgsToConfig()

	switch conf.typeOfConfig {
	case "args":
		logrus.Info("Use command-line arguments for configuration")
	case "env":
		logrus.Info("Use env variables for configuration")

		conf, err = envVarsToConfig()
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}
