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

	client := pb.NewTorrentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.List(ctx, &pb.Empty{})
	if err != nil {
		return errors.WithMessage(err, "Failed to list torrents")
	}

	for {
		to, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "Error while listing torrents")
		}

		torrentPrint(to)
	}

	return nil
}

// Add a new torrent
func Add(name string, magnet bool, directory string) error {
	// Set up a connection to the server.
	serverAddr := "localhost:" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "Did not connect")
	}
	defer conn.Close()

	client := pb.NewTorrentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var request pb.AddRequest
	request.Magnet = magnet
	if magnet {
		request.Name = name
	} else {
		fileAbsPath, err := filepath.Abs(name)
		if err != nil {
			return errors.WithMessage(err, "Could not resolve filepath")
		}
		request.Name = fileAbsPath
	}
	request.Directory = directory

	reply, err := client.Add(ctx, &request)
	if err != nil {
		return errors.WithMessage(err, "Failed to add torrent")
	}
	fmt.Printf("Added %d: %s %s\n", reply.GetId(), reply.GetName(), hex.EncodeToString(reply.GetInfoHash()))

	return nil
}

// Remove a managed torrent
func Remove(input string, isInfoHash bool) error {
	// Set up a connection to the server.
	serverAddr := "localhost:" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "Did not connect")
	}
	defer conn.Close()

	client := pb.NewTorrentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var id int
	var infoHash []byte
	var torrentRequest pb.TorrentRequest

	if isInfoHash {
		infoHash, err = hex.DecodeString(input)
		if err != nil {
			return errors.WithMessage(err, "Could not parse infohash")
		}
		torrentRequest.InfoHash = infoHash
	} else {
		id, err = strconv.Atoi(input)
		if err != nil {
			return errors.WithMessage(err, "Could not parse ID")
		}
		torrentRequest.Id = uint32(id)
	}

	reply, err := client.Remove(ctx, &torrentRequest)
	if err != nil {
		return errors.WithMessage(err, "Failed to remove torrent")
	}
	fmt.Printf("Removed %d: %s %s\n", reply.GetId(), reply.GetName(), hex.EncodeToString(reply.GetInfoHash()))

	return nil
}

// Start a torrent's download/upload
func Start(input string, isInfoHash bool) error {
	// Set up a connection to the server.
	serverAddr := "localhost:" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "Did not connect")
	}
	defer conn.Close()

	client := pb.NewTorrentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var id int
	var infoHash []byte
	var torrentRequest pb.TorrentRequest

	if isInfoHash {
		infoHash, err = hex.DecodeString(input)
		if err != nil {
			return errors.WithMessage(err, "Could not parse infohash")
		}
		torrentRequest.InfoHash = infoHash
	} else {
		id, err = strconv.Atoi(input)
		if err != nil {
			return errors.WithMessage(err, "Could not parse ID")
		}
		torrentRequest.Id = uint32(id)
	}

	reply, err := client.Start(ctx, &torrentRequest)
	if err != nil {
		return errors.WithMessage(err, "Failed to start torrent")
	}
	fmt.Printf("Started %d: %s %s\n", reply.GetId(), reply.GetName(), hex.EncodeToString(reply.GetInfoHash()))

	return nil
}

// Stop a torrent's download/upload
func Stop(input string, isInfoHash bool) error {
	// Set up a connection to the server.
	serverAddr := "localhost:" + strconv.Itoa(int(viper.GetViper().GetInt("server.port")))
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "Did not connect")
	}
	defer conn.Close()

	client := pb.NewTorrentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var id int
	var infoHash []byte
	var torrentRequest pb.TorrentRequest

	if isInfoHash {
		infoHash, err = hex.DecodeString(input)
		if err != nil {
			return errors.WithMessage(err, "Could not parse infohash")
		}
		torrentRequest.InfoHash = infoHash
	} else {
		id, err = strconv.Atoi(input)
		if err != nil {
			return errors.WithMessage(err, "Could not parse ID")
		}
		torrentRequest.Id = uint32(id)
	}

	reply, err := client.Stop(ctx, &torrentRequest)
	if err != nil {
		return errors.WithMessage(err, "Failed to start torrent")
	}
	fmt.Printf("Stopped %d: %s %s\n", reply.GetId(), reply.GetName(), hex.EncodeToString(reply.GetInfoHash()))

	return nil
}

func torrentPrint(to *pb.Torrent) {
	curr := to.GetTotalLength() - to.GetLeft()
	progress := float64(curr) / float64(to.GetTotalLength()) * 100

	fmt.Printf("%d: %-50s %s %s %s %s %s\n",
		to.Id,
		to.GetName(),
		fmt.Sprintf("infohash: %s", hex.EncodeToString(to.GetInfoHash())),
		fmt.Sprintf("progress: %.1f%%", progress),
		fmt.Sprintf("download: %s", ratePretty(to.GetDownRate())),
		fmt.Sprintf("upload: %s", ratePretty(to.GetUpRate())),
		fmt.Sprintf("state: %s", to.GetState().String()),
	)
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

// func sizePretty(size uint32) string {
// 	floatSize := float64(size)
// 	return ""
// }
