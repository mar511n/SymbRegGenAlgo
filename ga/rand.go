package ga

import (
	"math/rand"
	"sync"
	"time"
)

type lockedRand struct {
	mu sync.Mutex
	r  *rand.Rand
}

func (lr *lockedRand) Float64() float64 {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	return lr.r.Float64()
}

func (lr *lockedRand) Intn(n int) int {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	return lr.r.Intn(n)
}

var rnd = &lockedRand{
	r: rand.New(rand.NewSource(time.Now().UnixNano())),
}
