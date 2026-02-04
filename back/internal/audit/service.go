package audit

import (
	"context"
	"fmt"
	"sync"
)

// Service defines the audit logging interface.
type Service interface {
	Log(entry *AuditLog)
	Start()
	Stop()
}

type service struct {
	repo Repository
	ch   chan *AuditLog
	wg   sync.WaitGroup
}

// NewService creates a new audit service with a buffered channel for async writes.
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
		ch:   make(chan *AuditLog, 256),
	}
}

// Start begins the background goroutine that drains the audit channel.
func (s *service) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for entry := range s.ch {
			if err := s.repo.Create(context.Background(), entry); err != nil {
				fmt.Printf("audit: failed to write log: %v\n", err)
			}
		}
	}()
}

// Stop closes the channel and waits for all pending entries to be flushed.
func (s *service) Stop() {
	close(s.ch)
	s.wg.Wait()
}

// Log sends an audit entry to the background writer. Non-blocking; drops if buffer full.
func (s *service) Log(entry *AuditLog) {
	select {
	case s.ch <- entry:
	default:
		fmt.Println("audit: channel full, dropping log entry")
	}
}
