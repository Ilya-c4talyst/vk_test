package subpub

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestSubscribePublish проверяет базовый сценарий:
// - подписка на событие
// - публикация сообщения
// - получение сообщения подписчиком
func TestSubscribePublish(t *testing.T) {
	sp := NewSubPub()

	var wg sync.WaitGroup
	wg.Add(1)

	// Подписываемся на тему "test"
	_, err := sp.Subscribe("test", func(msg interface{}) {
		if msg != "hello" {
			t.Errorf("Ожидалось 'hello', получено %v", msg)
		}
		wg.Done()
	})

	if err != nil {
		t.Fatalf("Ошибка подписки: %v", err)
	}

	// Публикуем сообщение
	err = sp.Publish("test", "hello")
	if err != nil {
		t.Fatalf("Ошибка публикации: %v", err)
	}

	wg.Wait() // Ждем обработки
}

// TestUnsubscribe проверяет:
// - отписку от событий
// - что после отписки сообщения не приходят
func TestUnsubscribe(t *testing.T) {
	sp := NewSubPub()

	sub, err := sp.Subscribe("test", func(msg interface{}) {
		t.Error("Сообщение получено после отписки!")
	})

	if err != nil {
		t.Fatalf("Ошибка подписки: %v", err)
	}

	// Отписываемся
	sub.Unsubscribe()

	// Публикуем сообщение (не должно быть получено)
	err = sp.Publish("test", "hello")
	if err != nil {
		t.Fatalf("Ошибка публикации: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

// TestClose проверяет:
// - корректное завершение работы
// - невозможность публикации после закрытия
func TestClose(t *testing.T) {
	sp := NewSubPub()

	_, err := sp.Subscribe("test", func(msg interface{}) {
		t.Error("Сообщение получено после закрытия!")
	})

	if err != nil {
		t.Fatalf("Ошибка подписки: %v", err)
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Закрываем шину
	err = sp.Close(ctx)
	if err != nil {
		t.Fatalf("Ошибка закрытия: %v", err)
	}

	// Пытаемся опубликовать (должно вернуть ошибку)
	err = sp.Publish("test", "hello")
	if err == nil {
		t.Error("Ожидалась ошибка при публикации в закрытую шину")
	}
}

// TestSlowSubscriber проверяет:
// - работу с медленными подписчиками
// - что медленный подписчик не блокирует систему
func TestSlowSubscriber(t *testing.T) {
	sp := NewSubPub()

	var wg sync.WaitGroup
	wg.Add(2)

	// Быстрый подписчик
	_, err := sp.Subscribe("test", func(msg interface{}) {
		wg.Done()
	})

	if err != nil {
		t.Fatalf("Ошибка подписки: %v", err)
	}

	// Медленный подписчик
	_, err = sp.Subscribe("test", func(msg interface{}) {
		time.Sleep(200 * time.Millisecond)
		wg.Done()
	})

	if err != nil {
		t.Fatalf("Ошибка подписки: %v", err)
	}

	// Публикуем 2 сообщения
	sp.Publish("test", "hello")
	sp.Publish("test", "world")

	wg.Wait()
}

// TestMultipleSubscribers проверяет:
// - множественную подписку на один subject
// - что все подписчики получают сообщения
func TestMultipleSubscribers(t *testing.T) {
	sp := NewSubPub()
	count := 0
	expected := 2

	var wg sync.WaitGroup
	wg.Add(expected)

	handler := func(msg interface{}) {
		count++
		wg.Done()
	}

	// Первый подписчик
	_, err := sp.Subscribe("test", handler)
	if err != nil {
		t.Fatalf("Ошибка подписки: %v", err)
	}

	// Второй подписчик
	_, err = sp.Subscribe("test", handler)
	if err != nil {
		t.Fatalf("Ошибка подписки: %v", err)
	}

	// Публикуем 1 сообщение (должно быть получено двумя подписчиками)
	err = sp.Publish("test", "hello")
	if err != nil {
		t.Fatalf("Ошибка публикации: %v", err)
	}

	wg.Wait()

	if count != expected {
		t.Errorf("Ожидалось %d обработок, получено %d", expected, count)
	}
}

// TestConcurrentAccess проверяет:
// - работу в конкурентной среде
// - отсутствие гонок данных
func TestConcurrentAccess(t *testing.T) {
	sp := NewSubPub()
	var wg sync.WaitGroup

	// Запускаем 10 подписчиков
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := sp.Subscribe("test", func(msg interface{}) {})
			if err != nil {
				t.Errorf("Ошибка подписки: %v", err)
			}
		}()
	}

	// Запускаем 10 издателей
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sp.Publish("test", "hello")
			if err != nil {
				t.Errorf("Ошибка публикации: %v", err)
			}
		}()
	}

	wg.Wait()
}
