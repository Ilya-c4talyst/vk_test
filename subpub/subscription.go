package subpub

// subscription представляет подписку для управления
type subscription struct {
	sub     *subscriber
	subject string
	sp      *subPub
}

// Unsubscribe отменяет подписку
func (s *subscription) Unsubscribe() {
	s.sp.mu.Lock()
	defer s.sp.mu.Unlock()

	select {
	case <-s.sub.done: // Если канал уже закрыт, ничего не делаем
		return
	default:
	}

	subs := s.sp.subjects[s.subject]
	for i, sub := range subs {
		if sub == s.sub {
			subs = append(subs[:i], subs[i+1:]...) // Удаляем подписчика
			s.sp.subjects[s.subject] = subs
			break
		}
	}
	close(s.sub.done) // Закрываем канал, если он ещё не закрыт
}
