package main

import (
	"context"
	"fmt"
	pb "github.com/mdShakilHossainNsu2018/grpc-go-course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"os/signal"
)

var collection *mongo.Collection

type server struct {
	pb.UnimplementedBlogServiceServer
}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func (s *server) CreateBlog(ctx context.Context, req *pb.CreateBlogRequest) (*pb.CreateBlogResponse, error) {
	fmt.Println("Create new blog called")

	blog := req.GetBlog()
	data := blogItem{AuthorID: blog.GetAuthorId(), Title: blog.GetTitle(), Content: blog.GetContent()}

	res, err := collection.InsertOne(context.Background(), data)

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal Error: %v\n", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)

	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cann't convert to OID"))
	}

	return &pb.CreateBlogResponse{
		Blog: &pb.Blog{Id: oid.Hex(), AuthorId: blog.GetAuthorId(), Title: blog.GetTitle(), Content: blog.GetContent()},
	}, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting MongoDB")
	client, mongoErr := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))

	if mongoErr != nil {
		log.Fatalf("Error from mongo: %v\n", mongoErr)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Error while desconnecting mongo: %v\n", err)
		}
	}()

	collection = client.Database("mydb").Collection("blog")

	fmt.Println("Starting server")
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

	pb.RegisterBlogServiceServer(s, &srv)

	log.Printf("server listening at %v", lis.Addr())

	go func() {
		fmt.Println("Starting server.")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Faild to serve: %v\n", err)
		}
	}()

	waitCh := make(chan os.Signal, 1)
	signal.Notify(waitCh, os.Interrupt)
	<-waitCh
	fmt.Println("Stopping the server.")
	s.Stop()
	fmt.Println("Closing the listener.")
	lisErr := lis.Close()
	if lisErr != nil {
		log.Fatalf("Close listener: %v\n", lisErr)
	}
	fmt.Println("Successfully ended program.")

}
