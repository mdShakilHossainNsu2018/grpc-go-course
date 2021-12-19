package main

import (
	"context"
	"fmt"
	"github.com/mdShakilHossainNsu2018/grpc-go-course/blog/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
)

func main() {
	fmt.Println("Hello I am from blog client!!!")

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

	c := blogpb.NewBlogServiceClient(cc)

	blog := &blogpb.Blog{Content: "New content", Title: "Title", AuthorId: "1"}

	createBlog, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Unable to create blog: %v\n", err)
	}

	fmt.Printf("Blog Created: %v\n", createBlog)
}
