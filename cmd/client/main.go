// Тестовый клиент для отладки
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "github.com/Ilya-c4talyst/vk_test/api"
	cust_errors "github.com/Ilya-c4talyst/vk_test/custom_errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddress = "localhost:50051"
	testTimeout   = 20 * time.Second
)

func main() {
	// Настройка подключения к серверу
	conn, err := grpc.Dial(serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("Ошибка подключения к серверу: %s", err.Error())
	}
	defer conn.Close()

	client := pb.NewPubSubClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Обработка сигналов прерывания
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nПолучен сигнал завершения...")
		cancel()
	}()

	// Запуск тестов
	runTestSuite(ctx, client)
}

func runTestSuite(ctx context.Context, client pb.PubSubClient) {
	fmt.Println("\n=== Начало тестирования ===")
	defer fmt.Println("\n=== Тестирование завершено ===")

	tests := []struct {
		name string
		fn   func(context.Context, pb.PubSubClient)
	}{
		{"Одиночная подписка", testSingleSubscription},
		{"Множественные подписчики", testMultipleSubscribers},
	}

	for _, test := range tests {
		fmt.Printf("\n--- Запуск теста: %s ---\n", test.name)
		test.fn(ctx, client)
		fmt.Printf("--- Тест завершен: %s ---\n", test.name)
	}
}

func testSingleSubscription(ctx context.Context, client pb.PubSubClient) {
	const testKey = "single_test"
	subCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Подписка
	stream, err := client.Subscribe(subCtx, &pb.SubscribeRequest{Key: testKey})
	if err != nil {
		log.Printf("Ошибка подписки: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Обработка сообщений
	go func() {
		defer wg.Done()
		for {
			event, err := stream.Recv()
			if err != nil {
				if cust_errors.IsExpectedError(err) {
					fmt.Println("Корректное завершение работы")
					return
				}
				log.Printf("Критическая ошибка: %v", err)
				return
			}
			fmt.Printf("[Одиночная подписка] Получено: %s\n", event.Data)
		}
	}()

	// Публикация тестовых сообщений
	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("Сообщение %d", i)
		if _, err := client.Publish(ctx, &pb.PublishRequest{
			Key:  testKey,
			Data: msg,
		}); err != nil {
			log.Printf("Ошибка публикации: %s", err.Error())
		}
		time.Sleep(300 * time.Millisecond)
	}

	wg.Wait()
}

func testMultipleSubscribers(ctx context.Context, client pb.PubSubClient) {
	const (
		testKey  = "multi_test"
		subCount = 3
		pubCount = 5
	)

	var wg sync.WaitGroup

	// Запуск подписчиков
	for i := 0; i < subCount; i++ {
		wg.Add(1)
		go func(subID int) {
			defer wg.Done()
			subCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			stream, err := client.Subscribe(subCtx, &pb.SubscribeRequest{Key: testKey})
			if err != nil {
				log.Printf("[Подписчик %d] Ошибка подписки: %s", subID, err.Error())
				return
			}

			for {
				event, err := stream.Recv()
				if err != nil {
					if cust_errors.IsExpectedError(err) {
						fmt.Println("Корректное завершение работы")
						return
					}
					log.Printf("Критическая ошибка: %v", err)
					return
				}
				fmt.Printf("[Подписчик %d] Получено: %s\n", subID, event.Data)
			}
		}(i)
	}

	// Даем время на подключение подписчиков
	time.Sleep(1 * time.Second)

	// Публикация сообщений
	for i := 1; i <= pubCount; i++ {
		msg := fmt.Sprintf("Групповое сообщение %d", i)
		if _, err := client.Publish(ctx, &pb.PublishRequest{
			Key:  testKey,
			Data: msg,
		}); err != nil {
			log.Printf("Ошибка публикации: %s", err.Error())
		}
		time.Sleep(200 * time.Millisecond)
	}

	wg.Wait()
}
