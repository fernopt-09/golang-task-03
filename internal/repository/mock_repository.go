package repository

import (
	"context"
	"sync"
	"time"

	"github.com/kazah/golang-task-03/internal/domain"
)

// MockRateRepository keeps rates in memory, it is used in unit tests.
type MockRateRepository struct {
	mu      sync.RWMutex
	Rates   []domain.Rate
	SaveErr error
}

// NewMockRateRepository creates an empty in-memory repository.
func NewMockRateRepository() *MockRateRepository {
	return &MockRateRepository{
		Rates: make([]domain.Rate, 0),
	}
}

// Save stores the rate in memory. If SaveErr is set, it returns this error.
func (m *MockRateRepository) Save(_ context.Context, rate *domain.Rate) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SaveErr != nil {
		return m.SaveErr
	}

	rate.ID = int64(len(m.Rates) + 1)
	if rate.CreatedAt.IsZero() {
		rate.CreatedAt = time.Now()
	}
	m.Rates = append(m.Rates, *rate)
	return nil
}

// GetLatest returns the last saved rate.
func (m *MockRateRepository) GetLatest(_ context.Context) (*domain.Rate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Rates) == 0 {
		return nil, domain.ErrRateNotFound
	}
	latest := m.Rates[len(m.Rates)-1]
	return &latest, nil
}

// List returns saved rates with limit and offset.
func (m *MockRateRepository) List(_ context.Context, limit, offset int) ([]domain.Rate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if offset >= len(m.Rates) {
		return []domain.Rate{}, nil
	}

	end := offset + limit
	if end > len(m.Rates) {
		end = len(m.Rates)
	}

	result := make([]domain.Rate, end-offset)
	copy(result, m.Rates[offset:end])
	return result, nil
}

// Ping always works, there is no real database.
func (m *MockRateRepository) Ping(_ context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.SaveErr
}
