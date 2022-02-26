package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pb "grpc_stream_simplified/helloworld/helloworld"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedGreeter1Server
	reply []*pb.HelloReply
}

//simple grpc
func (s *server) SayHello1(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Recieved: %v", in.GetMessage())
	return &pb.HelloReply{Message: "Hello" + in.GetMessage()}, nil
}

//server-side streaming grpc
func (s *server) SayHello2(in *pb.HelloRequest, stream pb.Greeter1_SayHello2Server) error {
	log.Printf("\nThis is server-side streaming grpc(server recieving these messages")
	message := string(in.GetMessage())
	log.Printf("Recieved: %v", message)
	for _, rply := range s.reply {
		if err := stream.Send(rply); err != nil {
			return err
		}
	}
	return nil
}

//client-side streaming grpc
func (s *server) SayHello3(stream pb.Greeter1_SayHello3Server) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.HelloReply{Message: "Client-side streaming recieved"})
		}
		if err != nil {
			return err
		}
	}
}

func (s *server) SayHello4(stream pb.Greeter1_SayHello4Server) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		reply := "Reply is " + in.Message
		if err := stream.Send(&pb.HelloReply{Message: reply}); err != nil {
			return err
		}

	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeter1Server(s, newServer())
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to served: %v", err)
	}
}

func newServer() *server {
	s := &server{}
	if err := json.Unmarshal(data, &s.reply); err != nil {
		log.Fatalf("failed to load data: %v", err)
	}
	return s
}

var data = []byte(`[
	{"message": "number 1"},
	{"message": "number 2"},
	{"message": "number 3"}]`)
