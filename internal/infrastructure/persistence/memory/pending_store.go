// Package memory implements in-memory pending payment store.
package memory

import (
	"log"
	"sync"
	"time"

	"github.com/exernia/botjanweb/pkg/logger"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// PendingStore manages pending payments in memory (thread-safe, FIFO).
type PendingStore struct {
	mu       sync.RWMutex
	pending  map[int][]*entity.PendingPayment
	logger   *log.Logger
	stopChan chan struct{}
}

// NewPendingStore creates a new in-memory pending payment store.
func NewPendingStore() *PendingStore {
	return &PendingStore{
		pending:  make(map[int][]*entity.PendingPayment),
		logger:   logger.Payment,
		stopChan: make(chan struct{}),
	}
}

// Add registers a new pending payment.
func (s *PendingStore) Add(p *entity.PendingPayment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[p.Amount] = append(s.pending[p.Amount], p)
	s.logger.Printf("Pending ditambahkan: Rp%d | MsgID: %s", p.Amount, p.MessageID)
}

// Match finds and removes the oldest pending payment matching the amount (FIFO).
func (s *PendingStore) Match(amount int) *entity.PendingPayment {
	s.mu.Lock()
	defer s.mu.Unlock()

	payments := s.pending[amount]
	if len(payments) == 0 {
		return nil
	}

	matched := payments[0]
	if len(payments) == 1 {
		delete(s.pending, amount)
	} else {
		s.pending[amount] = payments[1:]
	}

	s.logger.Printf("Pending dicocokkan: Rp%d | MsgID: %s", amount, matched.MessageID)
	return matched
}

// Count returns total pending payments.
func (s *PendingStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, p := range s.pending {
		count += len(p)
	}
	return count
}

// StartCleanup runs cleanup at midnight WIB, removing payments older than 24h.
func (s *PendingStore) StartCleanup() {
	wib := time.FixedZone("WIB", 7*60*60)

	go func() {
		for {
			now := time.Now().In(wib)
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, wib)
			wait := time.Until(next)

			s.logger.Printf("Cleanup dijadwalkan: %s WIB (dalam %s)", next.Format("02-01-2006 15:04:05"), wait.Round(time.Minute))

			select {
			case <-time.After(wait):
				s.cleanup(24 * time.Hour)
			case <-s.stopChan:
				return
			}
		}
	}()
}

// StopCleanup stops the cleanup goroutine.
func (s *PendingStore) StopCleanup() {
	close(s.stopChan)
}

// Close is a no-op for in-memory store (implements PendingStorePort).
func (s *PendingStore) Close() error {
	return nil
}

func (s *PendingStore) cleanup(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for amount, payments := range s.pending {
		var kept []*entity.PendingPayment
		for _, p := range payments {
			if p.CreatedAt.After(cutoff) {
				kept = append(kept, p)
			} else {
				removed++
			}
		}
		if len(kept) == 0 {
			delete(s.pending, amount)
		} else {
			s.pending[amount] = kept
		}
	}

	if removed > 0 {
		s.logger.Printf("Cleanup: %d pending kadaluarsa dihapus", removed)
	}
}
