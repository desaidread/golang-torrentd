package main

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"torrentd/internal/rpc"
)

type mode int

const (
	modeList  mode = iota // таблица раздач, опрос ListTorrents раз в секунду
	modeAdd               // ввод пути к .torrent
	modeWatch             // стрим WatchProgress по одной раздаче
)

type model struct {
	client rpc.TorrentdClient
	addr   string

	mode     mode
	torrents []*rpc.TorrentInfo
	cursor   int
	err      error
	status   string // однострочная подсказка/результат последнего действия

	input string // буфер ввода пути в modeAdd

	watching *rpc.TorrentInfo   // последнее событие из стрима
	events   <-chan tea.Msg     // канал стрима
	cancel   context.CancelFunc // остановить стрим при выходе из экрана

	width int
}

func newModel(c rpc.TorrentdClient, addr string) model {
	return model{client: c, addr: addr, width: 80}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(listTorrents(m.client), tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tickMsg:
		// Опрашиваем список только когда он виден: в modeWatch данные
		// приходят стримом, в modeAdd обновление только мешало бы.
		if m.mode == modeList {
			return m, tea.Batch(listTorrents(m.client), tick())
		}
		return m, tick()

	case torrentsMsg:
		m.torrents = msg
		m.err = nil
		if m.cursor >= len(m.torrents) {
			m.cursor = max(0, len(m.torrents)-1)
		}
		return m, nil

	case addedMsg:
		m.status = "добавлено: " + msg.id
		return m, listTorrents(m.client)

	case errMsg:
		m.err = msg.err
		return m, nil

	case watchStartMsg:
		m.events, m.cancel = msg.events, msg.cancel
		return m, waitForEvent(m.events)

	case progressMsg:
		m.watching = msg.info
		return m, waitForEvent(m.events) // тянем следующее событие
	case watchDoneMsg:
		m.status = "стрим завершён"
		m.events = nil
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.stopWatch()
		return m, tea.Quit
	}

	switch m.mode {

	case modeAdd:
		switch msg.String() {
		case "esc":
			m.mode, m.input = modeList, ""
		case "enter":
			path := strings.TrimSpace(m.input)
			m.mode, m.input = modeList, ""
			if path != "" {
				return m, addTorrent(m.client, path)
			}
		case "backspace":
			if len(m.input) > 0 {
				r := []rune(m.input)
				m.input = string(r[:len(r)-1])
			}
		default:
			// Key.Text не пустой только для печатаемых символов —
			// стрелки и ctrl+что-то сюда не попадут.
			m.input += msg.Text
		}
		return m, nil

	case modeWatch:
		switch msg.String() {
		case "esc", "q":
			m.stopWatch()
			m.mode, m.watching = modeList, nil
			return m, listTorrents(m.client)
		}
		return m, nil

	default: // modeList
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.torrents)-1 {
				m.cursor++
			}
		case "a":
			m.mode, m.status = modeAdd, ""
		case "r":
			return m, listTorrents(m.client)
		case "enter":
			if t := m.current(); t != nil {
				m.mode, m.watching, m.status = modeWatch, t, ""
				return m, startWatch(m.client, t.Id)
			}
		}
		return m, nil
	}
}

func (m model) current() *rpc.TorrentInfo {
	if m.cursor < 0 || m.cursor >= len(m.torrents) {
		return nil
	}
	return m.torrents[m.cursor]
}

func (m *model) stopWatch() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.events = nil
}

// ---- отрисовка ----

func (m model) View() tea.View {
	var b strings.Builder

	fmt.Fprintf(&b, "torrentctl — %s\n\n", m.addr)

	switch m.mode {
	case modeAdd:
		b.WriteString("Путь к .torrent-файлу:\n\n")
		fmt.Fprintf(&b, "  %s█\n\n", m.input)
		b.WriteString("enter — добавить, esc — отмена\n")

	case modeWatch:
		if t := m.watching; t != nil {
			fmt.Fprintf(&b, "%s\n\n", t.Name)
			fmt.Fprintf(&b, "  %s\n\n", progressBar(t, 40))
			fmt.Fprintf(&b, "  статус: %s\n  кусков: %d / %d\n\n", t.Status, t.Downloaded, t.Total)
		}
		b.WriteString("esc — назад к списку\n")

	default:
		if len(m.torrents) == 0 {
			b.WriteString("  Раздач нет. Нажмите «a», чтобы добавить .torrent\n")
		}
		for i, t := range m.torrents {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}
			fmt.Fprintf(&b, "%s%-28s %s %s\n", cursor, truncate(t.Name, 28), progressBar(t, 20), t.Status)
		}
		b.WriteString("\n↑/↓ — выбор, enter — следить, a — добавить, r — обновить, q — выход\n")
	}

	if m.status != "" {
		fmt.Fprintf(&b, "\n%s\n", m.status)
	}
	if m.err != nil {
		fmt.Fprintf(&b, "\nОшибка: %v\n", m.err)
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	v.WindowTitle = "torrentctl"
	return v
}

func progressBar(t *rpc.TorrentInfo, width int) string {
	if t.Total <= 0 {
		return strings.Repeat("░", width) + "   ?%"
	}
	ratio := float64(t.Downloaded) / float64(t.Total)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	return fmt.Sprintf("%s%s %3.0f%%",
		strings.Repeat("█", filled),
		strings.Repeat("░", width-filled),
		ratio*100)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
