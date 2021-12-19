package main

import (
	"context"
	"fmt"
	"github.com/mdShakilHossainNsu2018/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("Hello I am client!!!")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}
	defer func(cc *grpc.ClientConn) {
		var err = cc.Close()
		if err != nil {

		}
	}(cc)

	c := calculatorpb.NewCalculatorServiceClient(cc)

	// doUnary(c)
	// doServerStream(c)
	// doClientStream(c)
	// doBIDIStream(c)

	doErrorUnary(c)
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting sqrt unary RPC")

	doErrorCall(c, 10)

	doErrorCall(c, -4)

}

func doErrorCall(c calculatorpb.CalculatorServiceClient, n int32) {
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: n})

	if err != nil {
		resErr, ok := status.FromError(err)

		if ok {
			fmt.Printf(resErr.Message())
			fmt.Println(resErr.Code())
			// fmt.Println(resErr.Details())
			return
		} else {
			log.Fatalf("Critical error: %v\n", err)
			return
		}
	}

	fmt.Printf("Result %v: %v\n", n, res.GetNumberRoot())
}

func doClientStream(c calculatorpb.CalculatorServiceClient) {
	requests := []*calculatorpb.ComputeAverageRequest{
		&calculatorpb.ComputeAverageRequest{Number: 1},
		&calculatorpb.ComputeAverageRequest{Number: 2},
		&calculatorpb.ComputeAverageRequest{Number: 3},
		&calculatorpb.ComputeAverageRequest{Number: 4},
	}
	res, err := c.ComputeAverage(context.Background())
	for _, item := range requests {

		if err != nil {
			log.Fatalf("Error while invoked Compute average:%v\n", err)
		}

		fmt.Printf("Sending %v\n", item)

		err2 := res.Send(item)
		time.Sleep(1000 * time.Millisecond)
		if err2 != nil {
			log.Fatalf("Error while sending data: %v\n", err2)
		}
	}
	recv, err := res.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while recive and close: %v\n", err)
	}

	fmt.Printf("AVG: %v\n", recv.Average)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.SumRequest{FirstNumber: 1, SecondNumber: 3}

	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling rpc: %v\n", err)

	}
	log.Printf("Response from Greet: %v\n", res.SumResult)
}

func doServerStream(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.PrimeNumberRequest{Number: 120}

	res, err := c.PrimeNumberDecomposition(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while calling rpc: %v\n", err)
	}

	for true {
		response, err := res.Recv()
		if err == io.EOF {
			fmt.Println("End of file")
			break
		}

		if err != nil {
			log.Fatalf("Error while reciving data: %v\n", err)
		}

		log.Println(response.Result)
	}
}

func doBIDIStream(c calculatorpb.CalculatorServiceClient) {
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while invoking FindMax: %v\n", err)
	}

	waitc := make(chan struct{})

	go func() {
		numbers := []int32{1, 4, 6, 1, 12, 54, 43, 99, 3, 5, 6, 7}

		for _, num := range numbers {
			err := stream.Send(&calculatorpb.FindMaximumRequest{Number: num})
			fmt.Printf("Sending: %v\n", num)
			if err != nil {
				log.Fatalf("Error while sending data: %v\n", err)
			}
			time.Sleep(1000 * time.Millisecond)
		}

		err := stream.CloseSend()
		if err != nil {
			log.Fatalf("Error while CloseSend: %v\n", err)
		}

	}()

	go func() {
		for {
			res, err := stream.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("Error while recv: %v\n", err)
			}
			maximum := res.GetMaximum()

			fmt.Printf("Current Max is: %v\n", maximum)
		}
		close(waitc)
	}()

	<-waitc
}
