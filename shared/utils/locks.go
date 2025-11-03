package utils

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"

	"github.com/rs/zerolog/log"
)

type AdvisoryLock struct {
	db     *sql.DB
	lockID int64
	key    string
}

func NewAdvisoryLock(db *sql.DB, key string) *AdvisoryLock {
	lockID := hashToLockID(key)
	return &AdvisoryLock{
		db:     db,
		lockID: lockID,
		key:    key,
	}
}

func hashToLockID(key string) int64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	hash := h.Sum64()
	return int64(hash & 0x7FFFFFFFFFFFFFFF)
}

func (l *AdvisoryLock) TryAcquire(ctx context.Context) (bool, error) {
	query := `SELECT pg_try_advisory_lock($1)`

	var acquired bool
	err := l.db.QueryRowContext(ctx, query, l.lockID).Scan(&acquired)
	if err != nil {
		return false, fmt.Errorf("failed to try acquire lock: %w", err)
	}

	if acquired {
		log.Info().
			Str("key", l.key).
			Int64("lock_id", l.lockID).
			Msg("Advisory lock acquired")
	} else {
		log.Warn().
			Str("key", l.key).
			Int64("lock_id", l.lockID).
			Msg("Advisory lock not acquired (already held)")
	}

	return acquired, nil
}

func (l *AdvisoryLock) Acquire(ctx context.Context) error {
	query := `SELECT pg_advisory_lock($1)`

	_, err := l.db.ExecContext(ctx, query, l.lockID)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	log.Info().
		Str("key", l.key).
		Int64("lock_id", l.lockID).
		Msg("Advisory lock acquired (blocking)")

	return nil
}

func (l *AdvisoryLock) Release(ctx context.Context) error {
	query := `SELECT pg_advisory_unlock($1)`

	var released bool
	err := l.db.QueryRowContext(ctx, query, l.lockID).Scan(&released)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if !released {
		log.Warn().
			Str("key", l.key).
			Int64("lock_id", l.lockID).
			Msg("Advisory lock not released (was not held)")
	} else {
		log.Info().
			Str("key", l.key).
			Int64("lock_id", l.lockID).
			Msg("Advisory lock released")
	}

	return nil
}

func (l *AdvisoryLock) TryAcquireShared(ctx context.Context) (bool, error) {
	query := `SELECT pg_try_advisory_lock_shared($1)`

	var acquired bool
	err := l.db.QueryRowContext(ctx, query, l.lockID).Scan(&acquired)
	if err != nil {
		return false, fmt.Errorf("failed to try acquire shared lock: %w", err)
	}

	if acquired {
		log.Info().
			Str("key", l.key).
			Int64("lock_id", l.lockID).
			Msg("Shared advisory lock acquired")
	}

	return acquired, nil
}

func (l *AdvisoryLock) ReleaseShared(ctx context.Context) error {
	query := `SELECT pg_advisory_unlock_shared($1)`

	var released bool
	err := l.db.QueryRowContext(ctx, query, l.lockID).Scan(&released)
	if err != nil {
		return fmt.Errorf("failed to release shared lock: %w", err)
	}

	if released {
		log.Info().
			Str("key", l.key).
			Int64("lock_id", l.lockID).
			Msg("Shared advisory lock released")
	}

	return nil
}

type LockScope struct {
	lock *AdvisoryLock
	ctx  context.Context
}

func WithLock(ctx context.Context, db *sql.DB, key string, fn func() error) error {
	lock := NewAdvisoryLock(db, key)

	acquired, err := lock.TryAcquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		return fmt.Errorf("lock already held for key: %s", key)
	}

	defer func() {
		if err := lock.Release(ctx); err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to release lock")
		}
	}()

	return fn()
}
