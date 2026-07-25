package rpcserver

import (
	"context"
	"fmt"
	"time"
	"torrentd/internal/download"
	"torrentd/internal/rpc"
)

type Server struct {
	rpc.UnimplementedTorrentdServer
	mgr *download.Manager
}

func New(mgr *download.Manager) *Server {
	return &Server{mgr: mgr}
}

func (s *Server) AddTorrent(ctx context.Context, req *rpc.AddTorrentRequest) (*rpc.AddTorrentResponse, error) {
	id, err := s.mgr.AddTorrent(req.Path)
	if err != nil {
		return nil, err
	}

	return &rpc.AddTorrentResponse{Id: id}, nil
}

func (s *Server) ListTorrents(ctx context.Context, _ *rpc.ListTorrentsRequest) (*rpc.ListTorrentsResponse, error) {
	infos := []*rpc.TorrentInfo{}
	for _, t := range s.mgr.List() {
		done, total, status := t.Progress()
		infos = append(infos, &rpc.TorrentInfo{
			Id:         t.Id,
			Name:       t.Name,
			Downloaded: int32(done),
			Total:      int32(total),
			Status:     status,
		})

	}
	return &rpc.ListTorrentsResponse{Torrents: infos}, nil
}

func (s *Server) WatchProgress(req *rpc.WatchProgressRequest, stream rpc.Torrentd_WatchProgressServer) error {
	for {
		t, ok := s.mgr.Get(req.Id)
		if !ok {
			return fmt.Errorf("torrent %s not found", req.Id)

		}

		done, total, status := t.Progress()

		if err := stream.Send(&rpc.TorrentInfo{
			Id:         req.Id,
			Name:       t.Name,
			Downloaded: int32(done),
			Total:      int32(total),
			Status:     status,
		}); err != nil {
			return err
		}
		if status == "done" || status == "error" {
			return nil
		}

		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-time.After(time.Second):
		}

	}
}
