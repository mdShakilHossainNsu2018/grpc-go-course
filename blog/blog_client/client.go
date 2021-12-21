package main

import (
	"context"
	"fmt"
	"github.com/mdShakilHossainNsu2018/grpc-go-course/blog/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
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

	// Create blog

	blog := &blogpb.Blog{Content: "New content", Title: "Title", AuthorId: "1"}

	createBlog, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Unable to create blog: %v\n", err)
	}

	fmt.Printf("Blog Created: %v\n", createBlog)

	// Read blog

	_, readErr := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "61c15413cd2f702c9362195f"})
	if readErr != nil {
		fmt.Printf("Error while Reading: %v\n", readErr)
	}

	readBlog, read2Err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: createBlog.GetBlog().GetId()})
	if read2Err != nil {
		fmt.Printf("Error while Reading blog; %v\n", read2Err)
	}

	fmt.Printf("Blog response: %v\n", readBlog)

	// update Blog

	updateBlog := &blogpb.Blog{Id: createBlog.GetBlog().GetId(), Content: "Updated Blog New content", Title: "Updated Title", AuthorId: "1"}

	response, updateErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: updateBlog})
	if updateErr != nil {
		fmt.Printf("Update blog error: %v\n", updateErr)
	}

	fmt.Printf("Updated blog: %v\n", response)

	// delete blog

	deleteRes, delErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: createBlog.GetBlog().GetId()})

	if delErr != nil {
		fmt.Printf("Unable to delete: %v\n", delErr)
	}

	fmt.Printf("Delete Res: %v\n", deleteRes)

	// List blog

	stream, listBlogErr := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})

	if listBlogErr != nil {
		log.Fatalf("Error in List blog: %v\n", listBlogErr)
	}

	for true {
		res, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("No more file")
			break
		}

		if err != nil {
			log.Fatalf("Error in list: %v\n", err)
		}

		fmt.Println(res.GetBlog())
	}
}
