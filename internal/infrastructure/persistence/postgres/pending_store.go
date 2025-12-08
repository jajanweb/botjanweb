// Package postgres implements PostgreSQL pending payment store.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/exernia/botjanweb/pkg/logger"

	"github.com/exernia/botjanweb/internal/domain/entity"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PendingStore manages pending payments in PostgreSQL (thread-safe, FIFO).
type PendingStore struct {
	db       *sql.DB
	logger   *log.Logger
	stopChan chan struct{}
	stopped  bool // Track if cleanup has been stopped
}

// NewPendingStore creates a new PostgreSQL pending payment store.
// Automatically creates the table if it doesn't exist.
func NewPendingStore(ctx context.Context, databaseURL string) (*PendingStore, error) {
	log := logger.Payment

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PendingStore{
		db:       db,
		logger:   log,
		stopChan: make(chan struct{}),
	}

	// Create table if not exists
	if err := store.createTable(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	log.Println("✅ PostgreSQL pending store initialized")
	return store, nil
}

// createTable creates the pending_payments table if it doesn't exist.
func (s *PendingStore) createTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS pending_payments (
			id SERIAL PRIMARY KEY,
			amount INTEGER NOT NULL,
			message_id TEXT NOT NULL,
			chat_id TEXT NOT NULL,
			sender_jid TEXT NOT NULL,
			sender_phone TEXT NOT NULL,
			original_message_id TEXT NOT NULL,
			is_self_qris BOOLEAN NOT NULL DEFAULT FALSE,
			group_notif_msg_id TEXT,
			produk TEXT NOT NULL,
			nama TEXT NOT NULL,
			email TEXT NOT NULL,
			family TEXT,
			deskripsi TEXT,
			kanal TEXT,
			akun TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		-- Index for faster matching by amount (FIFO order)
		CREATE INDEX IF NOT EXISTS idx_pending_amount_created 
		ON pending_payments(amount, created_at);
	`

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// Add registers a new pending payment.
func (s *PendingStore) Add(p *entity.PendingPayment) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO pending_payments (
			amount, message_id, chat_id, sender_jid, sender_phone,
			original_message_id, is_self_qris, group_notif_msg_id,
			produk, nama, email, family, deskripsi, kanal, akun, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := s.db.ExecContext(ctx, query,
		p.Amount, p.MessageID, p.ChatID, p.SenderJID, p.SenderPhone,
		p.OriginalMessageID, p.IsSelfQris, p.GroupNotifMsgID,
		p.Produk, p.Nama, p.Email, p.Family, p.Deskripsi, p.Kanal, p.Akun, p.CreatedAt,
	)

	if err != nil {
		s.logger.Printf("❌ Failed to add pending payment: %v", err)
		return
	}

	s.logger.Printf("Pending ditambahkan: Rp%d | MsgID: %s", p.Amount, p.MessageID)
}

// Match finds and removes the oldest pending payment matching the amount (FIFO).
func (s *PendingStore) Match(amount int) *entity.PendingPayment {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Printf("❌ Failed to begin transaction: %v", err)
		return nil
	}
	defer tx.Rollback()

	// Find oldest matching payment (FIFO) with row lock
	query := `
		SELECT id, amount, message_id, chat_id, sender_jid, sender_phone,
		       original_message_id, is_self_qris, group_notif_msg_id,
		       produk, nama, email, family, deskripsi, kanal, akun, created_at
		FROM pending_payments
		WHERE amount = $1
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	var p entity.PendingPayment
	var id int64

	err = tx.QueryRowContext(ctx, query, amount).Scan(
		&id, &p.Amount, &p.MessageID, &p.ChatID, &p.SenderJID, &p.SenderPhone,
		&p.OriginalMessageID, &p.IsSelfQris, &p.GroupNotifMsgID,
		&p.Produk, &p.Nama, &p.Email, &p.Family, &p.Deskripsi, &p.Kanal, &p.Akun, &p.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		s.logger.Printf("❌ Failed to query pending payment: %v", err)
		return nil
	}

	// Delete the matched payment
	_, err = tx.ExecContext(ctx, "DELETE FROM pending_payments WHERE id = $1", id)
	if err != nil {
		s.logger.Printf("❌ Failed to delete pending payment: %v", err)
		return nil
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Printf("❌ Failed to commit transaction: %v", err)
		return nil
	}

	s.logger.Printf("Pending dicocokkan: Rp%d | MsgID: %s", p.Amount, p.MessageID)
	return &p
}

// Count returns total pending payments.
func (s *PendingStore) Count() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pending_payments").Scan(&count)
	if err != nil {
		s.logger.Printf("❌ Failed to count pending payments: %v", err)
		return 0
	}

	return count
}

// StartCleanup runs cleanup at midnight WIB, removing payments older than 24h.
func (s *PendingStore) StartCleanup() {
	go func() {
		wib := time.FixedZone("WIB", 7*60*60)
		s.logger.Println("🧹 Cleanup scheduler started (midnight WIB)")

		for {
			// Calculate next midnight WIB
			now := time.Now().In(wib)
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, wib)
			duration := next.Sub(now)

			s.logger.Printf("⏰ Next cleanup: %s (in %v)", next.Format("2006-01-02 15:04:05"), duration.Round(time.Minute))

			select {
			case <-time.After(duration):
				s.cleanup(24 * time.Hour)
			case <-s.stopChan:
				s.logger.Println("🛑 Cleanup scheduler stopped")
				return
			}
		}
	}()
}

// StopCleanup stops the cleanup routine.
func (s *PendingStore) StopCleanup() {
	if !s.stopped {
		s.stopped = true
		close(s.stopChan)
		s.logger.Println("🛑 Cleanup scheduler stopped")
	}
}

// cleanup removes pending payments older than maxAge.
func (s *PendingStore) cleanup(maxAge time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cutoff := time.Now().Add(-maxAge)

	result, err := s.db.ExecContext(ctx,
		"DELETE FROM pending_payments WHERE created_at < $1",
		cutoff,
	)

	if err != nil {
		s.logger.Printf("❌ Cleanup failed: %v", err)
		return
	}

	count, _ := result.RowsAffected()
	s.logger.Printf("🧹 Cleanup complete: %d expired pending(s) removed", count)
}

// Close closes the database connection.
func (s *PendingStore) Close() error {
	s.StopCleanup()
	return s.db.Close()
}
