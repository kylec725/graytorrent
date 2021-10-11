package client

import (
	"context"
	"fmt"
	"log"
	"strconv"

	pb "github.com/kylec725/gray/rpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Add a new torrent
func Add() error {
	// Set up a connection to the server.
	serverAddr := "localhost:" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
		return err
	}
	defer conn.Close()
	client := pb.NewTorrentClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reply, err := client.Add(ctx, &pb.AddRequest{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
		return err
	}
	fmt.Println("response:", reply.Message)
	return nil
}
