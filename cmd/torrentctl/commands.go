package main

import (
	"context"
	"errors"
	"io"
	"time"

	tea "charm.land/bubbletea/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"torrentd/internal/rpc"
)

const rpcTimeout = 5 * time.Second

// ---- сообщения ----
//
// Правило Bubble Tea: Update никогда не ходит в сеть сам. Любой ввод-вывод —
// это Cmd (функция, возвращающая Msg), а результат прилетает обратно в Update.

type (
	torrentsMsg   []*rpc.TorrentInfo // ответ ListTorrents
	addedMsg      struct{ id string }
	errMsg        struct{ err error }
	tickMsg       time.Time
	progressMsg   struct{ info *rpc.TorrentInfo } // одно событие из стрима
	watchDoneMsg  struct{}                        // стрим закрылся
	watchStartMsg struct {                        // стрим открыт, вот из чего читать
		events <-chan tea.Msg
		cancel context.CancelFunc
	}
)

func (e errMsg) Error() string { return e.err.Error() }

// ---- унарные вызовы ----

func listTorrents(c rpc.TorrentdClient) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
		defer cancel()

		resp, err := c.ListTorrents(ctx, &rpc.ListTorrentsRequest{})
		if err != nil {
			return errMsg{err}
		}
		return torrentsMsg(resp.Torrents)
	}
}

func addTorrent(c rpc.TorrentdClient, path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
		defer cancel()

		resp, err := c.AddTorrent(ctx, &rpc.AddTorrentRequest{Path: path})
		if err != nil {
			return errMsg{err}
		}
		return addedMsg{id: resp.Id}
	}
}

// tick — периодический опрос списка. Tick шлёт одно сообщение, поэтому
// в Update мы каждый раз возвращаем его снова, чтобы получился цикл.
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// ---- серверный стрим ----
//
// Cmd возвращает ровно одно сообщение, а стрим отдаёт много. Канонический
// приём: горутина читает stream.Recv() и пишет в канал, а модель после каждого
// события снова запускает waitForEvent — «насос», который тянет по одному.

func startWatch(c rpc.TorrentdClient, id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())

		stream, err := c.WatchProgress(ctx, &rpc.WatchProgressRequest{Id: id})
		if err != nil {
			cancel()
			return errMsg{err}
		}

		events := make(chan tea.Msg)

		go func() {
			defer close(events)
			for {
				info, err := stream.Recv()
				if err != nil {
					// io.EOF — сервер штатно завершил стрим (status == done/error).
					// codes.Canceled — это мы сами вызвали cancel(), выходя из экрана.
					if errors.Is(err, io.EOF) || status.Code(err) == codes.Canceled {
						return
					}
					select {
					case events <- errMsg{err}:
					case <-ctx.Done():
					}
					return
				}
				// select с ctx.Done() обязателен: без него горутина повиснет
				// на записи в канал, если TUI уже перестал читать.
				select {
				case events <- progressMsg{info}:
				case <-ctx.Done():
					return
				}
			}
		}()

		return watchStartMsg{events: events, cancel: cancel}
	}
}

func waitForEvent(events <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-events
		if !ok {
			return watchDoneMsg{}
		}
		return msg
	}
}
