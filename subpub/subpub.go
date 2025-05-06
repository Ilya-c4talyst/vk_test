package subpub

import (
	"context"
	"errors"
	"sync"
)

// MessageHandler определяет функцию обработки сообщений
type MessageHandler func(msg interface{})

// Subscription представляет подписку на события
type Subscription interface {
	Unsubscribe()
}

// SubPub реализует шину событий
type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

// Реализация SubPub
type subPub struct {
	mu       sync.RWMutex
	subjects map[string][]*subscriber
	closed   bool
}

// NewSubPub создает новый экземпляр SubPub
func NewSubPub() SubPub {
	return &subPub{
		subjects: make(map[string][]*subscriber),
	}
}

// Subscribe добавляет подписчика на события
func (s *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("subpub is closed")
	}

	sub := &subscriber{
		handler:  cb,
		signal:   make(chan struct{}, 1),
		done:     make(chan struct{}),
		closedCh: make(chan struct{}),
	}

	s.subjects[subject] = append(s.subjects[subject], sub)
	go sub.run()

	return &subscription{sub: sub, subject: subject, sp: s}, nil
}

// Publish отправляет сообщение всем подписчикам
func (s *subPub) Publish(subject string, msg interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return errors.New("subpub is closed")
	}

	for _, sub := range s.subjects[subject] {
		sub.publish(msg)
	}

	return nil
}

// Close завершает работу шины событий
func (s *subPub) Close(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true

	allSubs := make([]*subscriber, 0)
	for subject, subs := range s.subjects {
		allSubs = append(allSubs, subs...)
		delete(s.subjects, subject)
	}
	s.mu.Unlock()

	for _, sub := range allSubs {
		sub.close()
	}

	var wg sync.WaitGroup
	wg.Add(len(allSubs))
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for _, sub := range allSubs {
		go func(s *subscriber) {
			defer wg.Done()
			select {
			case <-s.closedCh:
			case <-ctx.Done():
			}
		}(sub)
	}

	select {
	case <-doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
