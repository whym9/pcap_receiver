package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	uploadpb "pcap_receiver/proto/proto"
)

func main() {
	fileName := *flag.String("name", "lo.pcapng", "path to the file")
	addr := *flag.String("address", ":8080", "address of the GRPC server")
	flag.Parse()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	client := NewClient(conn)
	name, err := client.Upload(context.Background(), fileName)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(name)
}

type Client struct {
	client uploadpb.UploadServiceClient
}

func NewClient(conn grpc.ClientConnInterface) Client {
	return Client{
		client: uploadpb.NewUploadServiceClient(conn),
	}
}

func (c Client) Upload(ctx context.Context, file string) (string, error) {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer cancel()

	stream, err := c.client.Upload(ctx)
	if err != nil {
		return "", err
	}

	fil, err := os.Open(file)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024)
	count := 0
	for {
		num, err := fil.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if err := stream.Send(&uploadpb.UploadRequest{Chunk: buf[:num]}); err != nil {
			return "", err
		}
		count++
	}
	fmt.Println(count)

	res, err := stream.CloseAndRecv()
	if err != nil {
		return "", err
	}

	return res.GetName(), nil
}
