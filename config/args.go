/*
Copyright Scientific Ideas 2022. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"flag"
	"time"
)

func cliArgsToConfig() *Config {
	pollInterval := flag.String("poll-interval", "1000ms", "ping interval")
	pingTimeout := flag.String("ping-timeout", "1000ms", "ping request timeout")
	logLevel := flag.String("log", "info", "logs level")
	redisPass := flag.String("redis-pass", "", "Redis password")
	redisAddrs := flag.String("redis-addrs", "0.0.0.0:6379", "Redis nodes addresses")
	redisTLS := flag.Bool("redis-tls", false, "enable TLS for communication with Redis")
	redisRootCACerts := flag.String("redis-rootca-certs", "", "comma-separated root CA's certificates list for TLS with Redis")
	distributionNamespace := flag.String("distribution-namespace", "sys-matching-table", "key in storage where work distribution data is stored")
	serviceStorageNamespaces := flag.String("services-namespaces", "sys-robots-list,sys-parsers-lists", "keys in storage where services lists are stored")
	workUnitsStorageNamespace := flag.String("workunits-namespace", "sys-channels", "key in storage where work units list is stored")
	promPort := flag.Uint("prom-port", 9090, "Prometheus metrics port")
	kaTime := flag.Duration("ka-time", 10*time.Second, "KeepAlive time")
	kaTimeout := flag.Duration("ka-timeout", 20*time.Second, "KeepAlive timeout")
	kaPermitWithoutStream := flag.Bool("ka-permit-without-stream", false, "KeepAlive param: if true, client sends keepalive pings even with no active RPCs; if false, when there are no active RPCs, Time and Timeout will be ignored and no keepalive pings will be sent")
	typeOfConfig := flag.String("config-type", "args", "which type of config to use")
	flag.Parse()

	return &Config{
		LogLevel:                *logLevel,
		PollInterval:            *pollInterval,
		PingTimeout:             *pingTimeout,
		RedisAddrs:              *redisAddrs,
		RedisPass:               *redisPass,
		RedisTLS:                *redisTLS,
		RedisRootCACerts:        *redisRootCACerts,
		typeOfConfig:            *typeOfConfig,
		DistributionNamespace:   *distributionNamespace,
		ServiceStorageNamespace: *serviceStorageNamespaces,
		WorkunitsNamespace:      *workUnitsStorageNamespace,
		PromPort:                *promPort,
		KATime:                  *kaTime,
		KATimeout:               *kaTimeout,
		KAPermitWithoutStream:   *kaPermitWithoutStream,
	}
}
