/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"context"
	"log"
	"net"
	"strconv"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/scientificideas/distributor/pinger/grpc/proto"
	"google.golang.org/grpc"
)

type PingServer struct {
	pb.UnimplementedGRPCPingerServer
}

func (p *PingServer) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	return new(empty.Empty), nil
}

// Inject is a helper function for distributor managed services.
// It starts GRPC server that responds to liveness requests.
func Inject(onHost string, onPort uint) {
	lis, err := net.Listen("tcp", net.JoinHostPort(onHost, strconv.Itoa(int(onPort))))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pingServer := &PingServer{}
	pb.RegisterGRPCPingerServer(grpcServer, pingServer)

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf(err.Error())
	}
}
