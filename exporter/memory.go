package exporter

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/tomb.v2"

	"cabourotte/healthcheck"
)

// MemoryStore A store containing the latest healthchecks results
type MemoryStore struct {
	TTL     time.Duration
	Logger  *zap.Logger
	Results map[string]*healthcheck.Result

	Tick *time.Ticker
	t    tomb.Tomb
	lock sync.RWMutex
}

// NewMemoryStore creates a new memory store
func NewMemoryStore(logger *zap.Logger) *MemoryStore {
	return &MemoryStore{
		Logger:  logger,
		TTL:     time.Second * 120,
		Results: make(map[string]*healthcheck.Result),
	}
}

// Start starts the memory store
func (m *MemoryStore) Start() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Tick = time.NewTicker(time.Second * 30)
	m.t.Go(func() error {
		for {
			select {
			case <-m.Tick.C:
				m.purge()
			case <-m.t.Dying():
				return nil
			}
		}
	})
}

// Stop stops the memory store
func (m *MemoryStore) Stop() error {
	m.Tick.Stop()
	m.t.Kill(nil)
	m.t.Wait()
	return nil
}

// Add a new Result to the store
func (m *MemoryStore) add(result *healthcheck.Result) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Results[result.Name] = result
}

// purge the expired results
func (m *MemoryStore) purge() {
	m.lock.Lock()
	defer m.lock.Unlock()
	now := time.Now()
	for _, result := range m.Results {
		if now.After(result.Timestamp.Add(m.TTL)) {
			m.Logger.Info("expire healthcheck",
				zap.String("name", result.Name))
			delete(m.Results, result.Name)
		}
	}
}

func (m *MemoryStore) list() []healthcheck.Result {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result := make([]healthcheck.Result, 0, len(m.Results))
	for _, value := range m.Results {
		result = append(result, *value)
	}
	return result
}