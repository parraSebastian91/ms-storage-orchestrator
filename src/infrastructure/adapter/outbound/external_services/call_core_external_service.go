package external_services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/config"

	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ ports.IExternalService = (*ExternalService)(nil)

type ExternalService struct {
	client  *http.Client
	baseURL string
	logger  ports.ILoggerService
}

func NewExternalService(cfg config.ExternalServiceConfig, logger ports.ILoggerService) *ExternalService {
	return &ExternalService{
		baseURL: cfg.BaseURL,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (es *ExternalService) CallNotifyCoreService(ctx context.Context, payload AplicationModel.NotifyModel) error {
	url := fmt.Sprintf("%s/webhooks/notify", es.baseURL)
	es.logger.Info("Calling core service notify endpoint", map[string]interface{}{
		"url":           url,
		"correlationId": payload.CorrelationId,
		"gestor":        payload.Gestor,
	})
	return es.doWithRetry(ctx, http.MethodPut, url, payload)
}

// isRetryableStatus returns true for status codes worth retrying (5xx and 429 Too Many Requests).
func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// doWithRetry executes an HTTP request with exponential backoff retries.
// Retries on network errors and retryable status codes (5xx, 429).
// Non-retryable HTTP errors (4xx) return immediately without retrying.
func (es *ExternalService) doWithRetry(ctx context.Context, method, url string, body interface{}) error {
	const maxRetries = 3

	reqBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second // 1s, 2s
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := es.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: error calling service: %w", attempt+1, err)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return nil
		}

		respBodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		lastErr = fmt.Errorf("attempt %d: unexpected status code: %d, body: %s", attempt+1, resp.StatusCode, string(respBodyBytes))

		if !isRetryableStatus(resp.StatusCode) {
			// 4xx errors won't be fixed by retrying
			return lastErr
		}
	}

	return fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}
