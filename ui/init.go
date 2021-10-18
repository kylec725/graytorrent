package ui

import (
	"context"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	pb "github.com/kylec725/graytorrent/rpc"
	"google.golang.org/grpc"
)

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.updateTorrents())
	return tea.Batch(cmds...)
}

func (m model) updateTorrents() tea.Cmd { // NOTE: slow to update list all at once, use a stream to learn when new torrents are added (use separate goroutine that locks the torrents)
	return func() tea.Msg {
		conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
		if err != nil {
			// t.Rows = [][]string{{"graytorrent server cannot be reached"}}
			return errorMsg{err}
		}
		defer conn.Close()

		client := pb.NewTorrentServiceClient(conn)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		reply, err := client.List(ctx, &pb.Empty{})
		if err != nil {
			// t.Rows = [][]string{{"Failed to list torrents"}}
			return errorMsg{err}
		}

		// Sort the torrents
		sort.Slice(reply.Torrents, func(i, j int) bool {
			return reply.Torrents[i].Name < reply.Torrents[j].Name
		})

		return updateTorrentsMsg(reply.Torrents)
	}
}
