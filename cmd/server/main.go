package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ilya-c4talyst/vk_test/api"
	"github.com/Ilya-c4talyst/vk_test/envs"
	"github.com/Ilya-c4talyst/vk_test/internal/server"
	"github.com/Ilya-c4talyst/vk_test/subpub"
	"google.golang.org/grpc"
)

func main() {
	// Инициализация subpub
	sp := subpub.NewSubPub()

	// Создание gRPC сервера
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(func(
			ctx context.Context, req interface{},
			info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
		) (resp interface{}, err error) {
			log.Printf("Request: %s", info.FullMethod)
			return handler(ctx, req)
		}),
		grpc.StreamInterceptor(func(
			srv interface{}, ss grpc.ServerStream,
			info *grpc.StreamServerInfo, handler grpc.StreamHandler,
		) error {
			log.Printf("Начало стрима: %s", info.FullMethod)
			return handler(srv, ss)
		}),
	)
	// Регистрация сервиса
	api.RegisterPubSubServer(grpcServer, server.NewPubSubServer(sp))

	// Создание контекста
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Получение порта
	errEnvs := envs.LoadEnvs()
	if errEnvs != nil {
		// Вывод сообщения об ошибке
		log.Fatal("Ошибка инициализации ENV: ", errEnvs)
	} else {
		log.Println("Инициализация ENV прошла успешно")
	}

	// Запуск сервера
	lis, err := net.Listen("tcp", envs.ServerEnvs.PORT)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		log.Println("Сервер запущен")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Серверная ошибка: %v", err)
		}
	}()

	// Ожидание контекста
	<-ctx.Done()
	log.Println("Сервер выключен (контекст)")
}
