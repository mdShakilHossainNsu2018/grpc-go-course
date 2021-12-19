package main

import (
	"context"
	"fmt"
	"github.com/mdShakilHossainNsu2018/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("Hello I am client!!!")

	tsl := false

	opts := grpc.WithInsecure()
	if tsl {
		cartPath := "ssl/ca.crt"
		creds, sslErr := credentials.NewClientTLSFromFile(cartPath, "")
		if sslErr != nil {
			log.Fatalf("Unable to load ssl %v\n", sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}
	defer func(cc *grpc.ClientConn) {
		var err = cc.Close()
		if err != nil {

		}
	}(cc)

	c := greetpb.NewGreetServiceClient(cc)

	doUnary(c)

	// doServerStream(c)
	// doClientStream(c)
	// doBIDIStream(c)
	// doUnaryWithTimeout(c, 5*time.Second)
	// doUnaryWithTimeout(c, 1*time.Second)
}

func doUnary(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{Greeting: &greetpb.Greeting{FirstName: "Shakil", LastName: "Hossain"}}

	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling rpc: %v\n", err)

	}
	log.Printf("Response from Greet: %v\n", res.Result)
}

func doUnaryWithTimeout(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("doUnaryWithTimeout called")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req := &greetpb.GreetWithDeadLineRequest{Greeting: &greetpb.Greeting{FirstName: "Shakil", LastName: "Hossain"}}

	res, err := c.GreetWithDeadLine(ctx, req)
	if err != nil {

		resErr, ok := status.FromError(err)

		if ok {
			if resErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Request Timeout")
			} else {
				fmt.Printf("Unexpected Error: %v\n", resErr)
			}
		} else {
			log.Fatalf("Critical error: %v", err)
		}
		return

	}
	log.Printf("Response from Greet: %v\n", res.Result)
}

func doServerStream(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetManyTimesRequest{Greeting: &greetpb.Greeting{FirstName: "Shakil", LastName: "Hossain"}}

	res, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling rpc: %v\n", err)

	}

	for true {
		response, err := res.Recv()

		if err == io.EOF {
			log.Printf("Streaming closed.")
			break
		}
		if err != nil {
			log.Fatalf("Error while recv %v\n", err)
		}
		log.Printf(response.Result)
	}
}

func doClientStream(c greetpb.GreetServiceClient) {

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{FirstName: "Shakil", LastName: "Hossain"},
		},

		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{FirstName: "Shakil 2", LastName: "Hossain"},
		},

		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{FirstName: "Shakil 3", LastName: "Hossain"},
		},

		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{FirstName: "Shakil 4", LastName: "Hossain"},
		},
	}

	stream, err := c.LongGreet(context.Background())

	if err != nil {
		log.Fatalf("Error while recive called: %v\n", err)
	}

	for _, item := range requests {
		fmt.Printf("Sending req: %v\n", item)
		err := stream.Send(item)

		if err != nil {
			log.Fatalf("Error while sending %v\n", err)
		}
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Error while Close and recv: %v\n", err)
	}

	fmt.Printf("Response: %v\n", res)

}

func doBIDIStream(c greetpb.GreetServiceClient) {
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while calling GreetEveryone: %v\n", err)
	}

	waitClose := make(chan struct{})

	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{FirstName: "Shakil", LastName: "Hossain"},
		},

		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{FirstName: "Shakil 2", LastName: "Hossain"},
		},

		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{FirstName: "Shakil 3", LastName: "Hossain"},
		},

		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{FirstName: "Shakil 4", LastName: "Hossain"},
		},
	}

	go func() {
		for _, item := range requests {
			err := stream.Send(item)
			fmt.Printf("Sending: %v\n", item)
			if err != nil {
				log.Fatalf("Error while sending data: %v\n", err)
			}
			time.Sleep(1000 * time.Millisecond)
		}

		err := stream.CloseSend()
		if err != nil {
			log.Fatalf("Unable to close %v\n", err)
		}
	}()

	go func() {
		for true {
			recv, err := stream.Recv()

			if err == io.EOF {
				close(waitClose)
				fmt.Println("No more messages")
				break
			}

			if err != nil {
				close(waitClose)
				log.Fatalf("Error while reciving data: %v\n", err)
			}

			fmt.Printf("Result: %v\n", recv.GetResult())
		}
	}()

	<-waitClose
}
