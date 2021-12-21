package main

import (
	"context"
	"fmt"
	pb "github.com/mdShakilHossainNsu2018/grpc-go-course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
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

func (s *server) ReadBlog(ctx context.Context, req *pb.ReadBlogRequest) (*pb.ReadBlogResponse, error) {
	fmt.Println("Read Blog called.")
	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cann't pearse OID: %v\n", err))
	}

	blogObj := &blogItem{}

	filter := bson.M{"_id": oid}
	result := collection.FindOne(context.Background(), filter)

	ResultErr := result.Decode(blogObj)
	if ResultErr != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Couldn't Find object: %v\n", ResultErr))
	}

	return &pb.ReadBlogResponse{Blog: &pb.Blog{Id: blogObj.ID.Hex(), Title: blogObj.Title, Content: blogObj.Content, AuthorId: blogObj.Content}}, nil

}

func (s *server) UpdateBlog(ctx context.Context, req *pb.UpdateBlogRequest) (*pb.UpdateBlogResponse, error) {
	fmt.Println("UpdateBlog called")
	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Unable to parse : %v\n", err))
	}

	blogObj := &blogItem{}

	filter := bson.M{"_id": oid}
	result := collection.FindOne(context.Background(), filter)

	ResultErr := result.Decode(blogObj)
	if ResultErr != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Couldn't Find object: %v\n", ResultErr))
	}

	blogObj.AuthorID = blog.AuthorId
	blogObj.Title = blog.Title
	blogObj.Content = blog.Content

	_, updateErr := collection.ReplaceOne(context.Background(), filter, blogObj)

	if updateErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Couldn't update object: %v\n", updateErr))
	}

	return &pb.UpdateBlogResponse{Blog: dataToBlog(blogObj)}, nil
}

func dataToBlog(data *blogItem) *pb.Blog {
	return &pb.Blog{Id: data.ID.Hex(), Title: data.Title, Content: data.Content, AuthorId: data.AuthorID}
}

func (s *server) DeleteBlog(ctx context.Context, req *pb.DeleteBlogRequest) (*pb.DeleteBlogResponse, error) {
	fmt.Println("Delete Blog request")
	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cann't pearse OID: %v\n", err))
	}

	filter := bson.M{"_id": oid}
	result, delErr := collection.DeleteOne(context.Background(), filter)

	if delErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Unable to delete: %v\n", delErr))
	}

	if result.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Didn't found: %v\n", delErr))
	}

	return &pb.DeleteBlogResponse{BlogId: req.GetBlogId()}, nil
}

func (s *server) ListBlog(req *pb.ListBlogRequest, stream pb.BlogService_ListBlogServer) error {
	fmt.Println("ListBlog called")
	cursor, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknon error: %v\n", err))
	}

	defer func() {
		err := cursor.Close(context.Background())
		if err != nil {
			log.Fatalf("Unable to close cursor: %v\n", err)
		}
	}()

	for cursor.Next(context.Background()) {
		data := &blogItem{}
		err := cursor.Decode(data)

		fmt.Println(data)

		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Error while decoding: %v\n", err))
		}
		sendErr := stream.Send(&pb.ListBlogResponse{Blog: dataToBlog(data)})
		if sendErr != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Unknown error: %v\n", err))
		}
	}
	return nil
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
