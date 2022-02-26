package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	pb "grpc_stream_simplified/helloworld/helloworld"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultName = "World"

var (
	addr     = flag.String("addr", "localhost:50051", "the address to connect to")
	name     = flag.String("name", defaultName, "Name to greet")
	messages []*pb.HelloRequest
)

func main() {
	flag.Parse()
	messages = append(messages, &pb.HelloRequest{Message: "This is "})
	messages = append(messages, &pb.HelloRequest{Message: "client streaming "})
	messages = append(messages, &pb.HelloRequest{Message: "example"})
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeter1Client(conn)

	//simple grpc
	log.Printf("\nThis is simple grpc")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello1(ctx, &pb.HelloRequest{Message: *name})
	if err != nil {
		log.Fatalf("could not greet1: %v", err)

	}
	log.Printf("Greeting1: %s", r.GetMessage())

	//server-side streaming grpc
	log.Printf("\nThis is server-side streaming grpc")
	stream1, err := c.SayHello2(ctx, &pb.HelloRequest{Message: *name})
	if err != nil {
		log.Fatalf("%v.sayhello2 = _, %v", c, err)
	}
	for {
		reply, err := stream1.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.Sayhello2 = _, %v", c, err)
		}
		log.Printf("The reply is %v", reply.Message)
	}

	//client-side streaming grpc
	log.Printf("\nThis is client-side streaming grpc")
	stream2, err := c.SayHello3(ctx)
	if err != nil {
		log.Fatalf("%v.SayHello3(_) = _, %v", c, err)
	}
	for _, message := range messages {
		if err := stream2.Send(message); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream2, message, err)
		}
	}
	reply, err := stream2.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream2, err, nil)
	}
	log.Printf("The client-side streaming reply is %v", reply)

	//bi-directional streaming grpc
	log.Printf("\nThis is bi-directional streaming grpc")
	stream3, err := c.SayHello4(context.Background())
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream3.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("failed to recieved a reply : %v", err)
			}
			log.Printf("Got reply %v", in.Message)
		}
	}()
	for _, message := range messages {
		if err := stream3.Send(message); err != nil {
			log.Fatalf("failed to send a message: %v", err)
		}
	}
	stream3.CloseSend()
	<-waitc

}
