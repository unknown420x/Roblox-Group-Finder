package internal

import (
	"context"
	"sync"
	"time"
)

type AdaptiveLimiter struct {
	mu        sync.Mutex
	maxRPS    float64
	rps       float64
	next      time.Time
	successes int
}

func NewAdaptiveLimiter(maxRPS int) *AdaptiveLimiter {
	if maxRPS < 1 {
		maxRPS = 1
	}
	return &AdaptiveLimiter{maxRPS: float64(maxRPS), rps: float64(maxRPS)}
}

func (l *AdaptiveLimiter) CurrentRPS() float64 { l.mu.Lock(); defer l.mu.Unlock(); return l.rps }

func (l *AdaptiveLimiter) Wait(ctx context.Context) error {
	l.mu.Lock()
	now := time.Now()
	ready := l.next
	interval := time.Duration(float64(time.Second) / l.rps)
	if ready.Before(now) {
		ready = now
	}
	l.next = ready.Add(interval)
	wait := time.Until(ready)
	l.mu.Unlock()
	if wait <= 0 {
		return nil
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *AdaptiveLimiter) Backoff(delay time.Duration) {
	if delay < time.Second {
		delay = time.Second
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rps *= 0.5
	if l.rps < 1 {
		l.rps = 1
	}
	blocked := time.Now().Add(delay)
	if blocked.After(l.next) {
		l.next = blocked
	}
	l.successes = 0
}

func (l *AdaptiveLimiter) Success() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.successes++
	if l.successes < 50 {
		return
	}
	l.rps += 1
	if l.rps > l.maxRPS {
		l.rps = l.maxRPS
	}
	l.successes = 0
}
