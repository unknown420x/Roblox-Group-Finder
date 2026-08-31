package internal

import "testing"

func TestCycleGeneratorUnique(t *testing.T) {
	g := newCycleGenerator(100, 12345)
	seen := make(map[uint64]bool, 100)
	for i := 0; i < 100; i++ {
		v, ok := g.Next()
		if !ok {
			t.Fatal("generator ended early")
		}
		if v >= 100 {
			t.Fatalf("out of range: %d", v)
		}
		if seen[v] {
			t.Fatalf("duplicate %d", v)
		}
		seen[v] = true
	}
	for i := uint64(0); i < 100; i++ {
		if !seen[i] {
			t.Fatalf("missing %d", i)
		}
	}
}
