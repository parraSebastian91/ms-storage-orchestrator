package ports

import "context"

type IExternalService interface {
	CallNotifyCoreService(ctx context.Context, payload []byte) error
}
