package tonrates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/logx"
)

const (
	coingeckoURL = "https://api.coingecko.com/api/v3/simple/price?ids=the-open-network&vs_currencies=usd,eur,gbp,rub"
	cacheTTL     = 5 * time.Minute
)

type svc struct {
	client  *http.Client
	log     *slog.Logger
	mu      sync.RWMutex
	cached  *dto.TonRatesResponse
	fetchAt time.Time
}

func New(log *slog.Logger) *svc {
	log = log.With(logx.Service("TonRates"))
	return &svc{
		client: &http.Client{Timeout: 10 * time.Second},
		log:    log,
	}
}

func (s *svc) GetRates(ctx context.Context) (*dto.TonRatesResponse, error) {
	s.mu.RLock()
	if s.cached != nil && time.Since(s.fetchAt) < cacheTTL {
		defer s.mu.RUnlock()
		return s.cached, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cached != nil && time.Since(s.fetchAt) < cacheTTL {
		return s.cached, nil
	}

	rates, err := s.fetch(ctx)
	if err != nil {
		if s.cached != nil {
			s.log.Warn("coingecko fetch failed, returning stale cache", "error", err)
			return s.cached, nil
		}
		return nil, fmt.Errorf("fetch ton rates: %w", err)
	}

	s.cached = rates
	s.fetchAt = time.Now()
	return rates, nil
}

type coingeckoResponse struct {
	TON struct {
		USD float64 `json:"usd"`
		EUR float64 `json:"eur"`
		GBP float64 `json:"gbp"`
		RUB float64 `json:"rub"`
	} `json:"the-open-network"`
}

func (s *svc) fetch(ctx context.Context) (*dto.TonRatesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, coingeckoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var cg coingeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cg); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &dto.TonRatesResponse{
		USD: cg.TON.USD,
		EUR: cg.TON.EUR,
		GBP: cg.TON.GBP,
		RUB: cg.TON.RUB,
	}, nil
}
