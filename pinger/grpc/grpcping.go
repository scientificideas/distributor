/*
Copyright Scientific Ideas 2022. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/keepalive"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/scientificideas/distributor/pinger/grpc/client"
	"google.golang.org/grpc"
)

// Pinger is a gRPC implementation of Pinger interface that checks services liveness.
type Pinger struct {
	mu           sync.RWMutex
	connPool     map[string]*client.GRPCClient
	kaParameters keepalive.ClientParameters
}

func (p *Pinger) connAlreadyExistsInPool(url string) bool {
	var ok bool

	p.mu.RLock()
	_, ok = p.connPool[url]
	p.mu.RUnlock()

	return ok
}

func (p *Pinger) getConn(url string) *client.GRPCClient {
	var c *client.GRPCClient

	p.mu.RLock()
	c = p.connPool[url]
	p.mu.RUnlock()

	return c
}

func (p *Pinger) addConnToPool(url string, conn *client.GRPCClient) {
	p.mu.Lock()
	p.connPool[url] = conn
	p.mu.Unlock()
}

// NewPinger creates GRPCPinger instance.
func NewPinger(kaparameters keepalive.ClientParameters) *Pinger {
	return &Pinger{connPool: make(map[string]*client.GRPCClient)}
}

// Init creates connections to the all services and adds them to local pool.
func (p *Pinger) Init(urls ...string) error {
	for _, url := range urls {
		if !p.connAlreadyExistsInPool(url) {
			pingerClient, err := client.NewClient(url, p.kaParameters)
			if err != nil {
				return fmt.Errorf("failed to connect to client %s, %w", url, err)
			}

			p.addConnToPool(url, pingerClient)
		}
	}

	return nil
}

// Ping makes a gRPC call to the service to check it's liveness.
func (p *Pinger) Ping(ctx context.Context, url string) error {
	if p.connAlreadyExistsInPool(url) {
		c := p.getConn(url)
		_, err := c.Ping(ctx, &empty.Empty{}, grpc.WaitForReady(true))

		return err
	}

	c, err := client.NewClient(url, p.kaParameters)
	if err != nil {
		return fmt.Errorf("failed to connect to client %s, %w", url, err)
	}

	p.addConnToPool(url, c)
	_, err = c.Ping(ctx, &empty.Empty{}, grpc.WaitForReady(true))

	return err
}
