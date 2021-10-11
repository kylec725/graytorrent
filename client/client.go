package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strconv"

	pb "github.com/kylec725/graytorrent/rpc"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// List the currently managed torrents
func List() error {
	// Set up a connection to the server.
	serverAddr := "localhost:" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "Did not connect")
	}
	defer conn.Close()

	client := pb.NewTorrentClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.List(ctx, &pb.ListRequest{})
	if err != nil {
		return errors.WithMessage(err, "Failed to list torrents")
	}
	for {
		torrentInfo, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "Error while listing torrents")
		}

		curr := torrentInfo.GetTotalLength() - torrentInfo.GetLeft()
		progress := float64(curr) / float64(torrentInfo.GetTotalLength())

		fmt.Printf("%-50s %s %s %s %s %s\n",
			torrentInfo.GetName(),
			fmt.Sprintf("infohash: %s", hex.EncodeToString(torrentInfo.GetInfoHash())),
			fmt.Sprintf("progress: %.1f%%", progress),
			fmt.Sprintf("download: %s", ratePretty(torrentInfo.GetDownRate())),
			fmt.Sprintf("upload: %s", ratePretty(torrentInfo.GetUpRate())),
			fmt.Sprintf("state: %s", torrentInfo.GetState().String()),
		)
	}
	return nil
}

// Add a new torrent
func Add(file string) error {
	// Set up a connection to the server.
	serverAddr := "localhost:" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "Did not connect")
	}
	defer conn.Close()

	client := pb.NewTorrentClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileAbsPath, err := filepath.Abs(file)
	if err != nil {
		return errors.WithMessage(err, "Could not resolve filepath")
	}

	reply, err := client.Add(ctx, &pb.AddRequest{File: fileAbsPath})
	if err != nil {
		return errors.WithMessage(err, "Failed to add torrent")
	}
	fmt.Printf("Added %s %s\n", reply.GetName(), hex.EncodeToString(reply.GetInfoHash()))
	return nil
}

func ratePretty(rate uint32) string {
	floatRate := float64(rate)
	suffix := "B/s"
	if floatRate > 1024 {
		floatRate /= 1024
		suffix = "KiB/s"
	}
	if floatRate > 1024 {
		floatRate /= 1024
		suffix = "MiB/s"
	}
	return fmt.Sprintf("%.2f "+suffix, floatRate)
}
