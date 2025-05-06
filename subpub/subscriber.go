package subpub

import "sync"

// subscriber представляет подписчика с очередью сообщений
type subscriber struct {
	handler  MessageHandler
	mu       sync.Mutex
	queue    []interface{}
	signal   chan struct{}
	done     chan struct{}
	closedCh chan struct{}
}

// publish добавляет сообщение в очередь подписчика
func (s *subscriber) publish(msg interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return
	default:
	}

	s.queue = append(s.queue, msg)
	if len(s.queue) == 1 {
		select {
		case s.signal <- struct{}{}:
		default:
		}
	}
}

// run запускает обработчик сообщений подписчика
func (s *subscriber) run() {
	defer close(s.closedCh)

	for {
		select {
		case <-s.signal:
			s.processQueue()
		case <-s.done:
			s.processQueue()
			return
		}
	}
}

// processQueue обрабатывает сообщения из очереди
func (s *subscriber) processQueue() {
	for {
		s.mu.Lock()
		if len(s.queue) == 0 {
			s.mu.Unlock()
			return
		}
		msg := s.queue[0]
		s.queue = s.queue[1:]
		s.mu.Unlock()

		s.handler(msg)
	}
}

// close завершает работу подписчика
func (s *subscriber) close() {
	select {
	case <-s.done:
	default:
		close(s.done)
	}
}
