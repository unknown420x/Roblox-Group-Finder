package internal

// cycleGenerator visits every value in [0,n) exactly once per cycle
// when step and n are coprime. It requires O(1) memory.
type cycleGenerator struct {
	start   uint64
	current uint64
	step    uint64
	n       uint64
}

func newCycleGenerator(n, seed uint64) cycleGenerator {
	if n == 0 {
		return cycleGenerator{}
	}
	start := seed % n
	step := (seed<<1 | 1) % n
	if step == 0 {
		step = 1
	}
	for gcd(step, n) != 1 {
		step++
		if step >= n {
			step = 1
		}
	}
	return cycleGenerator{start: start, current: start, step: step, n: n}
}

func (g *cycleGenerator) Next() (uint64, bool) {
	if g.n == 0 {
		return 0, false
	}
	v := g.current
	g.current = (g.current + g.step) % g.n
	return v, true
}

func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
