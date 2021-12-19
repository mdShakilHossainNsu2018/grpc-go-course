package main

import (
	"context"
	"fmt"
	pb "github.com/mdShakilHossainNsu2018/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
	"time"
)

type server struct {
	pb.UnimplementedCalculatorServiceServer
}

// Sum implements
func (s *server) Sum(ctx context.Context, req *pb.SumRequest) (*pb.SumResponse, error) {
	fmt.Printf("Recived Sum RPC: %v\n", req)
	firstNumber := req.GetFirstNumber()
	secondNumber := req.GetSecondNumber()

	result := firstNumber + secondNumber

	res := &pb.SumResponse{
		SumResult: result,
	}
	return res, nil
}

func (s *server) PrimeNumberDecomposition(req *pb.PrimeNumberRequest, stream pb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Println("Starting Prime Calculator.")
	//k = 2
	//N = 210
	//while N > 1:
	//if N % k == 0:   // if k evenly divides into N
	//print k      // this is a factor
	//N = N / k    // divide N by k so that we have the rest of the number left.
	//else:
	//k = k + 1

	var K int32 = 2
	N := req.GetNumber()
	for N > 1 {
		if N%K == 0 {
			res := &pb.PrimeNumberResponse{Result: K}
			err := stream.Send(res)
			if err != nil {
				log.Fatalf("Error while sending: %v\n", err)
			}
			N = N / K
		} else {
			K = K + 1
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (s *server) ComputeAverage(stream pb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("ComputeAverage Invoked %v\n", stream)
	var result float64 = 0

	var num int32 = 0

	var count int32 = 0

	for {
		res, err := stream.Recv()

		if err == io.EOF {
			result = float64(num) / float64(count)
			return stream.SendAndClose(&pb.ComputeAverageResponse{Average: result})
		}

		if err != nil {
			log.Fatalf("Error while stream reaciving: %v\n", err)
		}

		count++

		num += res.GetNumber()
	}
	return nil

}

func (s *server) FindMaximum(stream pb.CalculatorService_FindMaximumServer) error {
	fmt.Println("FindMaximum called")

	maximum := int32(0)

	for {
		res, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("NO more item")
			return nil
		}

		if err != nil {
			log.Fatalf("Cann't recive %v\n", err)
		}

		num := res.GetNumber()
		fmt.Println(num)

		if maximum <= num {
			maximum = num
			err := stream.Send(&pb.FindMaximumResponse{Maximum: maximum})
			if err != nil {
				log.Fatalf("Error while sending %v\n", err)
				return err
			}
		}

	}

}

func (s *server) SquareRoot(c context.Context, req *pb.SquareRootRequest) (*pb.SquareRootResponse, error) {
	fmt.Println("Received SquareRoot RPC ")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v\n", number),
		)
	}
	return &pb.SquareRootResponse{NumberRoot: math.Sqrt(float64(number))}, nil
}

func main() {
	fmt.Println("Calculator Server")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Unable to listen: %v\n", err)
	}
	s := grpc.NewServer()

	var srv server

	pb.RegisterCalculatorServiceServer(s, &srv)

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Faild to serve: %v\n", err)
	}
}
