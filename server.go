package main

import (
	"context"

	pb "github.com/kylec725/graytorrent/rpc"
	viper "github.com/spf13/viper"
)

func (server *torrentServer) Connect(ctx context.Context, in *pb.ConnectRequest) (*pb.ConnectReply, error) {
	if in.GetKey() == viper.GetString("server.key") {

	}

	return nil, nil
}
