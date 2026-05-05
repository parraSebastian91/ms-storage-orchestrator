package external_services

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/config"

	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ ports.IExternalService = (*ExternalService)(nil)

type ExternalService struct {
	client  *http.Client
	baseURL string
}

func NewExternalService(cfg config.ExternalServiceConfig) *ExternalService {
	return &ExternalService{
		baseURL: cfg.BaseURL,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (es *ExternalService) CallNotifyCoreService(ctx context.Context, payload []byte) error {
	url := fmt.Sprintf("%s/notify", es.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := es.client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling notify service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
