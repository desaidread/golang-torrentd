package main

import (
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"torrentd/internal/download"
	"torrentd/internal/rpc"
	"torrentd/internal/rpcserver"
)

func main() {
	// 1. движок
	mgr, err := download.NewManager()
	if err != nil {
		log.Fatal("manager:", err)
	}

	// 2. слушаем TCP-порт для gRPC
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("listen:", err)
	}

	// 3. создаём gRPC-сервер и регистрируем наш сервис
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	rpc.RegisterTorrentdServer(grpcServer, rpcserver.New(mgr))

	// 4. graceful shutdown по Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Println("останавливаюсь...")
		grpcServer.GracefulStop() // перестать принимать, дать текущим RPC доработать
		mgr.Shutdown()            // дождаться закачек
		os.Exit(0)
	}()

	// 5. запуск — блокирует
	log.Println("gRPC слушает на :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("serve:", err)
	}
}
