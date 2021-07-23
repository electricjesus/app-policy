package main

import (
	"net"

	"github.com/projectcalico/felix/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	outputQueueLen = 100
)

var (
	_ *logrus.Entry
	_ proto.PolicySyncServer = (*nullPolicySyncServer)(nil)

	updates = make(chan proto.ToDataplane, outputQueueLen)

	server *grpc.Server
	opts   []grpc.ServerOption
)

func main() {
	lis, err := net.Listen("unix", "/tmp/policysync.sock")
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(opts...)

	policySyncServer := &nullPolicySyncServer{}
	policySyncServer.register(grpcServer)
	logrus.Info("listening to /tmp/policysync.sock")
	grpcServer.Serve(lis)
}

type nullPolicySyncServer struct{}

func (s *nullPolicySyncServer) Sync(sr *proto.SyncRequest, stream proto.PolicySync_SyncServer) error {
	logrus.Info("sync request: ", sr)
	for update := range updates {
		err := stream.Send(&update)
		if err != nil {
			logrus.WithError(err).Warn("Failed to send update to policy sync client")
			return err
		}
	}
	return nil
}

func (s *nullPolicySyncServer) register(g *grpc.Server) {
	proto.RegisterPolicySyncServer(g, s)
}
