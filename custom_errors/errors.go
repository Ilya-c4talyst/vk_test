package errors

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Кастомный ловец ошибок
func IsExpectedError(err error) bool {
	if err == nil {
		return false
	}
	// Проверка стандартных ошибок контекста
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}
	// Проверка gRPC статусов
	st, ok := status.FromError(err)
	if ok {
		return st.Code() == codes.Canceled || st.Code() == codes.DeadlineExceeded
	}
	return false
}
