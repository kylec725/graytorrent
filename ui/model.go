package ui

import pb "github.com/kylec725/graytorrent/rpc"

type model struct {
	torrents []*pb.Torrent
	cursor   int
	selected map[int]struct{}
}

func initialModel() model {
	return model{
		torrents: make([]*pb.Torrent, 0),
		selected: make(map[int]struct{}), // indicates which choices are selected
	}
}
