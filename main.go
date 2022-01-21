/*
Copyright Scientific Ideas 2022. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/keepalive"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/scientificideas/distributor/config"
	grpcping "github.com/scientificideas/distributor/pinger/grpc"
	"github.com/scientificideas/distributor/storage"
	"github.com/sirupsen/logrus"
)

var Version = "undefined"

func main() {
	configuration, err := config.GetConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	// expose Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	// expose app version
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte(Version))
	})

	// start REST server
	go func() {
		logrus.Error(http.ListenAndServe(fmt.Sprintf(":%d", configuration.PromPort), nil))
	}()

	lvl, err := logrus.ParseLevel(configuration.LogLevel)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.SetLevel(lvl)

	if configuration.KATime == 0 {
		configuration.KATime = 10 * time.Second
	}
	if configuration.KATimeout == 0 {
		configuration.KATimeout = 20 * time.Second
	}
	p := grpcping.NewPinger(keepalive.ClientParameters{
		Time:                configuration.KATime,    // default infinity
		Timeout:             configuration.KATimeout, // default 20s
		PermitWithoutStream: configuration.KAPermitWithoutStream,
	})

	pingTimeout, err := time.ParseDuration(configuration.PingTimeout)
	if err != nil {
		logrus.Fatal(err)
	}

	pollInterval, err := time.ParseDuration(configuration.PingTimeout)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("connecting to Redis...")

	storageInstance, err := storage.NewRedis(
		configuration.RedisPass,
		strings.Split(configuration.RedisAddrs, ","),
		configuration.RedisTLS,
		strings.Split(configuration.RedisRootCACerts, ","),
	)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("successfully connected to Redis")

	for _, serviceNamespace := range strings.Split(configuration.ServiceStorageNamespace, ",") {
		distributor, err := NewDistributor(
			configuration.DistributionNamespace,
			configuration.WorkunitsNamespace,
			serviceNamespace,
			p,
			WithStorage(storageInstance),
			WithTransport(&Transport{
				PingTimeout:  pingTimeout,
				PollInterval: pollInterval,
			}),
		)
		if err != nil {
			logrus.Fatal(err)
		}

		errorsChan := make(chan error, 100)

		go distributor.Run(errorsChan)
		go func() {
			for err := range errorsChan {
				logrus.Warn(err)
			}
		}()
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		os.Interrupt,
		syscall.SIGTERM,
	)

	s := <-signalChan
	logrus.Infof("Got signal: %s", s.String())
	logrus.Info("shutdown")
}
