package download

import (
	"sync"
)

type Torrent struct {
	Name       string
	Total      int
	mu         sync.Mutex
	downloaded int
	status     string
}

func (t *Torrent) Progress() (downloaded, total int, status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.downloaded, t.Total, t.status
}

func (t *Torrent) pieceDone() {
	t.mu.Lock()
	t.downloaded++
	t.mu.Unlock()
}

func (t *Torrent) setStatus(s string) {
	t.mu.Lock()
	t.status = s
	t.mu.Unlock()

}
