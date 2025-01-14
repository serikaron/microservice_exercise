package main

import (
	"context"
	"echo/proto"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

//go:generate protoc -I../proto --go_out=plugins=grpc,paths=source_relative:../proto/ echo.proto

const (
	port = ":55555"
)

type EchoService struct {
}

func (es *EchoService) Echo(_ context.Context, in *proto.EchoReq) (*proto.EchoRsp, error) {
	log.Println("req msg: ", in.Msg)
	serverName := "echo-service"
	msg := fmt.Sprintf("Echo from service [%s]: %s", serverName, in.Msg)
	return &proto.EchoRsp{Msg: msg}, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	proto.RegisterEchoServer(s, &EchoService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
