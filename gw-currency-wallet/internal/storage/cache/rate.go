package cache

import (
	"sync"
	"time"
)

type Pair struct {
	from string
	to   string
}

type Item struct {
	rate      float32
	expiresAt time.Time
}

type RateCache struct {
	data map[Pair]Item
	mu   sync.RWMutex
	ttl  time.Duration
}

func NewRateCache(ttl time.Duration) *RateCache {
	return &RateCache{
		data: make(map[Pair]Item),
		ttl:  ttl,
	}
}

func (c *RateCache) Set(from string, to string, rate float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[Pair{from, to}] = Item{
		rate:      rate,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *RateCache) Get(from string, to string) (float32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.data[Pair{from, to}]
	if !ok {
		return 0, false
	}

	if time.Now().After(item.expiresAt) {
		return 0, false
	}

	return item.rate, true
}
