package cmd

import (
	"context"
	"github.com/youpipe/go-youPipe/pbs"
	"github.com/youpipe/go-youPipe/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type cmdService struct{}

var defaultRes = &pbs.CommonResponse{Msg: "success"}

func startCmdService() {

	var address = "127.0.0.1:" + utils.CmdServicePort
	l, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("failed to start command server:->", err)
	}

	cmdServer := grpc.NewServer()
	pbs.RegisterCmdServiceServer(cmdServer, &cmdService{})

	reflection.Register(cmdServer)
	if err := cmdServer.Serve(l); err != nil {
		logger.Fatalf("failed to serve command :%v", err)
	}
}

func DialToCmdService() pbs.CmdServiceClient {

	var address = "127.0.0.1:" + utils.CmdServicePort
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("cmd dial to server err:->", err)
	}

	client := pbs.NewCmdServiceClient(conn)

	return client
}

func (s *cmdService) ShowNodeInfo(ctx context.Context, request *pbs.EmptyRequest) (*pbs.CommonResponse, error) {
	return defaultRes, nil
}
