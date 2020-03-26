package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"mse/chat/pkg"
	"mse/chat/proto"
	"net"
)

//const (
//	port = ":12345"
//)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	defer log.Println("service exit")

	host := flag.String("chat-service-host", "", "chat service host")
	port := flag.Uint("chat-service-port", 0, "chat service port")
	rdsHost := flag.String("redis-host", "redis", "redis host")
	rdsPort := flag.Uint("redis-port", 3697, "redis port")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("listening at addr -", addr)

	rdsAddr := fmt.Sprintf("%s:%d", *rdsHost, *rdsPort)
	rdsPS := pkg.NewRedisPubSub(rdsAddr, "notify")
	s := grpc.NewServer()
	cs := NewChatService(&ChatNotifier{}, rdsPS)
	defer cs.Close()
	proto.RegisterChatServer(s, cs)

	done := make(chan bool)

	go func() {
		defer close(done)

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		log.Println("service stop")
	}()

	cs.Run(done)
}