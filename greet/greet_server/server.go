package main

import (
	"context"
	"fmt"
	pb "github.com/mdShakilHossainNsu2018/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type server struct {
	pb.UnimplementedGreetServiceServer
}

// Greet implements Greet.GreeterServer
func (s *server) Greet(ctx context.Context, req *pb.GreetRequest) (*pb.GreetResponse, error) {
	firstName := req.Greeting.GetFirstName()

	str := "Hello " + firstName

	res := &pb.GreetResponse{
		Result: str,
	}
	return res, nil
}

func (s *server) GreetWithDeadLine(ctx context.Context, req *pb.GreetWithDeadLineRequest) (*pb.GreetWithDeadLineResponse, error) {
	fmt.Println("GreetWithDeadLine Called")

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("Request canceled")

			return nil, status.Error(codes.Canceled, "Request Canceled")
		}
		time.Sleep(1 * time.Second)
	}

	firstName := req.GetGreeting().GetFirstName()

	str := "Hello " + firstName + "\n"

	res := &pb.GreetWithDeadLineResponse{
		Result: str,
	}
	return res, nil
}

func (s *server) GreetManyTimes(req *pb.GreetManyTimesRequest, stream pb.GreetService_GreetManyTimesServer) error {

	fmt.Println("Starting sending stream")
	firstName := req.GetGreeting().GetFirstName()

	for i := 0; i <= 10; i++ {
		result := "Hello " + firstName + " iter: " + strconv.Itoa(i)
		res := &pb.GreetManyTimesResponse{

			Result: result,
		}
		err := stream.Send(res)
		if err != nil {
			log.Fatalf("Couldn't Send stream: %v\n", err)
			return err
		}

		time.Sleep(1000 * time.Millisecond)
	}
	fmt.Println("Streaming ended.")
	return nil
}

func (s *server) LongGreet(stream pb.GreetService_LongGreetServer) error {
	fmt.Printf("LongGreat Invoked: %v\n", stream)
	result := ""

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(
				&pb.LongGreetResponse{
					Result: result,
				},
			)

		}
		if err != nil {
			log.Fatalf("Error while reciving data")
		}
		firstName := res.GetGreeting().GetFirstName()
		result += firstName + "!!!"
	}
	return nil
}

func (s *server) GreetEveryone(stream pb.GreetService_GreetEveryoneServer) error {
	fmt.Println("GreetEveryone Invoked")

	for {
		res, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("Error while reading client stream: %v\n", err)
			return err
		}

		firstName := res.GetGreeting().GetFirstName()

		result := "Hello " + firstName + "!"

		err2 := stream.Send(&pb.GreetEveryoneResponse{Result: result})

		if err2 != nil {
			log.Fatalf("Error while sendind: %v\n", err2)
		}
	}

}

func main() {
	fmt.Println("Hello world!!!")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Unable to listen: %v\n", err)
	}

	tsl := false

	opts := []grpc.ServerOption{}
	if tsl {
		cartFilePath := "ssl/server.crt"
		keyFilePath := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(cartFilePath, keyFilePath)

		if sslErr != nil {
			log.Fatalf("Feaild to load SSL: %v\n", sslErr)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)

	var srv server

	pb.RegisterGreetServiceServer(s, &srv)

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Faild to serve: %v\n", err)
	}
}
