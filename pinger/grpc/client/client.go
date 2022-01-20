/*
Copyright Scientific Ideas 2022. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"google.golang.org/grpc/keepalive"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/scientificideas/distributor/pinger/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

type PingCall func(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)

type GRPCClient struct {
	// cc     grpc.ClientConnInterface
	client pb.GRPCPingerClient
	conn   *grpc.ClientConn
	URL    string
}

func PingerClient(target string, opts ...grpc.DialOption) (*GRPCClient, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{client: pb.NewGRPCPingerClient(conn), conn: conn}, nil
}

// func DialRetry(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
// 	var (
// 		err          error
// 		conn         *grpc.ClientConn
// 		initialDelay = 50
// 		multiplier   = 2
// 		n            = 5
// 	)
//
// 	conn, err = grpc.Dial(target, opts...)
// 	for err != nil && n > 0 {
// 		initialDelay *= multiplier
// 		conn, err = grpc.Dial(target, opts...)
// 		time.Sleep(time.Duration(initialDelay) * time.Millisecond)
// 		n--
// 	}
//
// 	return conn, err
// }

func (c *GRPCClient) Ping(ctx context.Context, _ *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	n := 15

	_, err := c.client.Ping(ctx, &empty.Empty{}, opts...)
	for err != nil && n > 0 {
		time.Sleep(100 * time.Millisecond)
		n--

		_, err = c.client.Ping(ctx, &empty.Empty{})
	}

	return nil, err
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

func NewClient(url string, keepaliveParams keepalive.ClientParameters) (*GRPCClient, error) {
	connectParams := grpc.ConnectParams{
		MinConnectTimeout: 1 * time.Second,
		Backoff: backoff.Config{
			BaseDelay:  5 * time.Millisecond,
			Multiplier: 1.1,
			Jitter:     0,
			MaxDelay:   1000 * time.Millisecond,
		},
	}

	return PingerClient(url, grpc.WithInsecure(), grpc.WithConnectParams(connectParams), grpc.WithKeepaliveParams(keepaliveParams))
}
