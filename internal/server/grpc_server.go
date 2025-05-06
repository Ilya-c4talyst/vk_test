package server

import (
	"context"
	"log"

	"github.com/Ilya-c4talyst/vk_test/api"
	"github.com/Ilya-c4talyst/vk_test/subpub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Структура для сервера
type PubSubServer struct {
	api.UnimplementedPubSubServer
	pubsub subpub.SubPub
}

// Конструктор для сервера
func NewPubSubServer(pubsub subpub.SubPub) *PubSubServer {
	return &PubSubServer{pubsub: pubsub}
}

// Реализация ручки Subscribe
func (s *PubSubServer) Subscribe(req *api.SubscribeRequest, stream api.PubSub_SubscribeServer) error {
	sub, err := s.pubsub.Subscribe(req.Key, func(msg interface{}) {
		if err := stream.Send(&api.Event{Data: msg.(string)}); err != nil {
			// Логика внутренней функции (заглушка)
			log.Println("Ошибка обработки сообщения: ", err)
		}
	})
	if err != nil {
		log.Println("Ошибка подписки: ", err)
		return status.Error(codes.Internal, "Ошибка подписки")
	}
	log.Println("Успешная подписка")
	defer sub.Unsubscribe()

	<-stream.Context().Done()
	return nil
}

// Реализация ручки Publish
func (s *PubSubServer) Publish(ctx context.Context, req *api.PublishRequest) (*emptypb.Empty, error) {
	if err := s.pubsub.Publish(req.Key, req.Data); err != nil {
		log.Println("Ошибка публикации: ", err)
		return nil, status.Error(codes.Internal, "Ошибка публикации")
	}
	log.Println("Успешная публикация")
	return &emptypb.Empty{}, nil
}
