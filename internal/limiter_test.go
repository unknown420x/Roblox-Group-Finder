package internal

import (
	"context"
	"testing"
	"time"
)

func TestLimiterInitialRPS(t *testing.T) {
	l := NewAdaptiveLimiter(20)
	if got := l.CurrentRPS(); got != 20 {
		t.Fatalf("got %.2f", got)
	}
}
func TestLimiterBackoff(t *testing.T) {
	l := NewAdaptiveLimiter(20)
	l.Backoff(time.Millisecond)
	if got := l.CurrentRPS(); got != 10 {
		t.Fatalf("got %.2f", got)
	}
}
func TestLimiterLowerBound(t *testing.T) {
	l := NewAdaptiveLimiter(2)
	for i := 0; i < 10; i++ {
		l.Backoff(time.Millisecond)
	}
	if got := l.CurrentRPS(); got < 1 {
		t.Fatalf("got %.2f", got)
	}
}
func TestLimiterDoesNotExceedMaximum(t *testing.T) {
	l := NewAdaptiveLimiter(5)
	for i := 0; i < 1000; i++ {
		l.Success()
	}
	if got := l.CurrentRPS(); got > 5 {
		t.Fatalf("got %.2f", got)
	}
}
func TestLimiterCancellation(t *testing.T) {
	l := NewAdaptiveLimiter(1)
	l.mu.Lock()
	l.next = time.Now().Add(time.Second)
	l.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.Wait(ctx); err != context.Canceled {
		t.Fatalf("err=%v", err)
	}
}
func TestLimiterBackoffDelay(t *testing.T) {
	l := NewAdaptiveLimiter(20)
	l.Backoff(100 * time.Millisecond)
	start := time.Now()
	if err := l.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d < 90*time.Millisecond {
		t.Fatalf("waited %s", d)
	}
}
