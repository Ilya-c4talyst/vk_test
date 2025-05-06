package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ilya-c4talyst/vk_test/subpub"
)

func main() {
	sp := subpub.NewSubPub()

	// Подписчик 1 (новости)
	sub1, err := sp.Subscribe("news", func(msg interface{}) {
		fmt.Printf("[Subscriber 1] Новость: %v\n", msg)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		fmt.Println("Отписываем Subscriber 1")
		sub1.Unsubscribe()
	}()

	// Подписчик 2 (новости)
	sub2, err := sp.Subscribe("news", func(msg interface{}) {
		fmt.Printf("[Subscriber 2] Новость: %v\n", msg)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		fmt.Println("Отписываем Subscriber 2")
		sub2.Unsubscribe()
	}()

	// Подписчик 3 (обновления)
	sub3, err := sp.Subscribe("updates", func(msg interface{}) {
		fmt.Printf("[Subscriber 3] Обновление: %v\n", msg)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		fmt.Println("Отписываем Subscriber 3")
		sub3.Unsubscribe()
	}()

	// Публикуем сообщения
	fmt.Println("Публикуем сообщения...")
	sp.Publish("news", "Go 1.20 released!")
	sp.Publish("updates", "Server maintenance scheduled")
	sp.Publish("news", "GopherCon 2023 announced")

	// Даём время на обработку
	time.Sleep(500 * time.Millisecond)

	// Медленный подписчик
	sub4, err := sp.Subscribe("slow", func(msg interface{}) {
		time.Sleep(1 * time.Second)
		fmt.Printf("[Slow Subscriber] Обработано: %v\n", msg)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		fmt.Println("Отписываем Slow Subscriber")
		sub4.Unsubscribe()
	}()

	sp.Publish("slow", "Сообщение 1")
	sp.Publish("slow", "Сообщение 2")

	// Graceful shutdown
	fmt.Println("Завершаем работу...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sp.Close(ctx); err != nil {
		log.Printf("Ошибка закрытия: %v\n", err)
	} else {
		fmt.Println("Шина закрыта корректно")
	}
}
