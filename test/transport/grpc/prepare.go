package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/test/transport/grpc/pb"
	kitgrpc "github.com/einouqo/ext-kit/transport/grpc"
)

func prepareTestGRPC(address string) (client *ClientBinding, tidy func(), err error) {
	srvTidy, err := prepareServer(address)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to prepare server: %+v", err)
	}
	client, cTidy, err := prepareClient(address)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to prepare client: %+v", err)
	}
	return client, func() {
		cTidy()
		srvTidy()
	}, nil
}

func prepareClientPB(address string) (client pb.EchoClient, tidy func(), err error) {
	cc, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to Dial: %+v", err)
	}

	client = pb.NewEchoClient(cc)

	return client, func() { _ = cc.Close() }, nil
}

func prepareClient(address string, opts ...kitgrpc.ClientOption) (client *ClientBinding, tidy func(), err error) {
	cc, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to Dial: %+v", err)
	}

	client = NewClientBinding(cc, opts...)

	return client, func() { _ = cc.Close() }, nil
}

func prepareServer(address string, opts ...kitgrpc.ServerOption) (tidy func(), err error) {
	var (
		server = grpc.NewServer()
		echo   = service.NewEcho()
	)

	sc, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("unable to listen: %+v", err)
	}

	pb.RegisterEchoServer(server, NewServerBinding(echo, opts...))
	go server.Serve(sc)

	return func() {
		server.GracefulStop()
		_ = sc.Close()
	}, nil
}
