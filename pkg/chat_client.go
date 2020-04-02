package pkg

import (
	"context"
	"io"
	"log"
	"mse/proto"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func (c *ChatClient) unaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	opts = append(opts, grpc.PerRPCCredentials(oauth.NewOauthAccess(&oauth2.Token{
		//AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6Ik1hcnJ5IiwiZXhwIjo4NjQwMH0.h7SvqoYRlXGTh8Qjc-PgZ34iukcveYXMRqGi9eBYec4",
		AccessToken: c.token,
	})))
	//start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	//end := time.Now()
	//logger("RPC: %s, start time: %s, end time: %s, err: %v", method, start.Format("Basic"), end.Format(time.RFC3339), err)
	return err
}

func (c *ChatClient) UpdateToken(t string) {
	c.token = t
}

type ChatClient struct {
	conn   *grpc.ClientConn
	client proto.ChatClient
	ctx    context.Context
	token  string
}

func NewChatClient(addr string, pemFile string) *ChatClient {
	log.Println("try connect to chat-service:", addr)

	creds, err := credentials.NewClientTLSFromFile(pemFile, "serika-server")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	cc := &ChatClient{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(creds), grpc.WithBlock(), grpc.WithUnaryInterceptor(cc.unaryInterceptor))
	//conn, err := grpc.DialContext(ctx, addr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		cancel()
		log.Fatalln("grpc.Dial failed:", err)
	}

	//return &ChatClient{
	//	conn:   conn,
	//	client: proto.NewChatClient(conn),
	//	ctx:    ctx,
	//}

	cc.conn = conn
	cc.client = proto.NewChatClient(conn)
	return cc
}

func (cc *ChatClient) Listen(done chan bool) (chan string, error) {
	stream, err := cc.client.Listen(context.Background(), &proto.ListenReq{})
	if err != nil {
		log.Println("ChatClient.Listen failed:", err)
		return nil, err
	}

	c := make(chan string)
	go func() {
		defer close(c)

		for {
			select {
			case <-done:
				return
			default:
				inf, err := stream.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					log.Printf("ChatClient.Listen, stream.Recv failed %v", err)
					return
				}
				c <- inf.Msg
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	return c, nil
}

func (cc *ChatClient) Say(req *proto.SayReq) error {
	rsp, err := cc.client.Say(context.Background(), req)
	if err != nil {
		log.Println("ChatClient.Say failled:", err)
	} else {
		log.Println("ChatClient.Say rsp.Msg:", rsp.Msg)
	}
	return err
}
