package ports

import "context"

type IMessagePublisher interface {
	Publish(ctx context.Context, exchange string, message []byte) error
	PublishJson(ctx context.Context, exchange string, message interface{}) error
}

type IWorkPublisher interface {
	EnqueueWork(ctx context.Context, eventName string, payload []byte) error
}
