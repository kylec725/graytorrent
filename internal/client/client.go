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

var (
	address = "localhost:" + strconv.Itoa(int(viper.GetInt("server.port")))
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewTorrentClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reply, err := client.List(ctx, &pb.ListRequest{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	fmt.Println(reply)
}
