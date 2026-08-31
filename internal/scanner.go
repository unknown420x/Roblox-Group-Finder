package internal

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Scanner struct {
	client  *RobloxClient
	webhook *WebhookClient
	limiter *AdaptiveLimiter

	workers   int
	batchSize int
	minID     int
	maxID     int
	unique    bool

	checked     atomic.Uint64
	requests    atomic.Uint64
	hits        atomic.Uint64
	rateLimited atomic.Uint64
	errors      atomic.Uint64

	stateMu sync.RWMutex
	state   State
}

func NewScanner(client *RobloxClient, webhook *WebhookClient, cfg Config) *Scanner {
	return &Scanner{client: client, webhook: webhook, limiter: NewAdaptiveLimiter(cfg.RPS), workers: cfg.Workers, batchSize: cfg.BatchSize, minID: cfg.MinID, maxID: cfg.MaxID, unique: cfg.Unique}
}

func (s *Scanner) Run(ctx context.Context) {
	if s.unique {
		s.runUnique(ctx)
		return
	}
	s.runRandom(ctx)
}

func (s *Scanner) runUnique(ctx context.Context) {
	total := uint64(s.maxID - s.minID + 1)
	state := s.resumeState(total)
	if state.Done {
		fmt.Println("All IDs in the configured range have been scanned.")
		return
	}

	g := cycleGenerator{start: state.Start, current: state.Current, step: state.Step, n: total}
	jobs := make(chan []int, s.workers*2)
	var workers sync.WaitGroup
	workers.Add(s.workers)
	for i := 0; i < s.workers; i++ {
		go func() {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case batch, ok := <-jobs:
					if !ok {
						return
					}
					if err := s.limiter.Wait(ctx); err != nil {
						return
					}
					s.check(ctx, batch)
				}
			}
		}()
	}

	var emittedSinceSave uint64
	for state.Emitted < total {
		batch := make([]int, 0, s.batchSize)
		for len(batch) < s.batchSize && state.Emitted < total {
			id, _ := g.Next()
			batch = append(batch, s.minID+int(id))
			gCount := uint64(len(batch))
			state.Emitted += gCount
		}
		state.Current = g.current
		emittedSinceSave += uint64(len(batch))

		select {
		case jobs <- batch:
		case <-ctx.Done():
			_ = s.saveState(state)
			close(jobs)
			workers.Wait()
			return
		}

		if emittedSinceSave >= uint64(s.batchSize*20) {
			_ = s.saveState(state)
			emittedSinceSave = 0
		}
	}

	state.Done = true
	_ = s.saveState(state)
	close(jobs)
	workers.Wait()
}

func (s *Scanner) runRandom(ctx context.Context) {
	jobs := make(chan []int, s.workers*2)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobs)
		rng := newRandSource()
		for {
			batch := make([]int, s.batchSize)
			for i := range batch {
				batch[i] = s.minID + int(rng.next()%uint64(s.maxID-s.minID+1))
			}
			select {
			case jobs <- batch:
			case <-ctx.Done():
				return
			}
		}
	}()
	wg.Add(s.workers)
	for i := 0; i < s.workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case batch, ok := <-jobs:
					if !ok {
						return
					}
					if err := s.limiter.Wait(ctx); err != nil {
						return
					}
					s.check(ctx, batch)
				}
			}
		}()
	}
	wg.Wait()
}

type fastRand struct{ x uint64 }

func newRandSource() *fastRand {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return &fastRand{x: uint64(time.Now().UnixNano()) | 1}
	}
	x := binary.LittleEndian.Uint64(b[:]) | 1
	return &fastRand{x: x}
}
func (r *fastRand) next() uint64 {
	x := r.x
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	r.x = x
	return x
}

func (s *Scanner) resumeState(total uint64) State {
	state, err := LoadState()
	if err == nil && state.MinID == s.minID && state.MaxID == s.maxID && state.Step != 0 && state.Start < total && state.Current < total && state.Emitted <= total {
		return state
	}
	g := newCycleGenerator(total, newRandSource().next())
	return State{MinID: s.minID, MaxID: s.maxID, Start: g.start, Current: g.current, Step: g.step}
}

func (s *Scanner) saveState(state State) error {
	s.stateMu.Lock()
	s.state = state
	s.stateMu.Unlock()
	return SaveState(state)
}

func (s *Scanner) check(ctx context.Context, ids []int) {
	groups, status, retry, err := s.client.GetGroups(ctx, ids)
	s.requests.Add(1)
	if err != nil {
		s.errors.Add(1)
		return
	}
	if status == 429 {
		s.rateLimited.Add(1)
		s.limiter.Backoff(retry)
		return
	}
	if status < 200 || status >= 300 {
		s.errors.Add(1)
		return
	}
	s.limiter.Success()
	s.checked.Add(uint64(len(ids)))
	for _, group := range groups {
		if group.Owner != nil || !group.PublicEntryAllowed || group.IsLocked {
			continue
		}
		s.hits.Add(1)
		fmt.Printf("\n[+] HIT %d | %s\n", group.ID, group.Name)
		if s.webhook != nil {
			s.webhook.Queue(group.ID)
		}
	}
}

func (s *Scanner) Snapshot() Stats {
	return Stats{Requests: s.requests.Load(), Checked: s.checked.Load(), Hits: s.hits.Load(), RateLimited: s.rateLimited.Load(), Errors: s.errors.Load()}
}
