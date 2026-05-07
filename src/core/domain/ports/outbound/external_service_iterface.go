package ports

import (
	"context"

	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
)

type IExternalService interface {
	CallNotifyCoreService(ctx context.Context, payload AplicationModel.NotifyModel) error
}
