package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"torrentd/internal/rpc"
)

func main() {
	addr := os.Getenv("TORRENTD_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	// grpc.NewClient не подключается сразу — соединение ленивое, поднимется
	// при первом вызове. Это то, что нужно TUI: не блокируем старт интерфейса.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "не удалось создать клиента:", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := rpc.NewTorrentdClient(conn)

	if _, err := tea.NewProgram(newModel(client, addr)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ошибка TUI:", err)
		os.Exit(1)
	}
}
