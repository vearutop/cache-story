package cached

import (
	"context"
	"sync"
	"time"

	"github.com/bool64/cache"
	"github.com/bool64/stats"
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

func NewGreetingMaker(upstream greeting.Maker, cache *cache.FailoverOf[string]) *GreetingMaker {
	return &GreetingMaker{
		upstream: upstream,
		cache:    cache,
	}
}

type GreetingMaker struct {
	upstream greeting.Maker
	cache    *cache.FailoverOf[string]
}

func (g *GreetingMaker) GreetingMaker() greeting.Maker {
	return g
}

func (g *GreetingMaker) Hello(ctx context.Context, params greeting.Params) (string, error) {
	key := []byte(params.Name + params.Locale)
	return g.cache.Get(ctx, key, func(ctx context.Context) (string, error) {
		return g.upstream.Hello(ctx, params)
	})
}

type NaiveGreetingMaker struct {
	mu       sync.RWMutex
	ttl      time.Duration
	data     map[greeting.Params]greetingEntry
	upstream greeting.Maker
	stats    stats.Tracker
}

type greetingEntry struct {
	value   string
	expires time.Time
}

func NewNaiveGreetingMaker(upstream greeting.Maker, ttl time.Duration, stats stats.Tracker) *NaiveGreetingMaker {
	return &NaiveGreetingMaker{
		ttl:      ttl,
		data:     map[greeting.Params]greetingEntry{},
		upstream: upstream,
		stats:    stats,
	}
}

func (g *NaiveGreetingMaker) GreetingMaker() greeting.Maker {
	return g
}

func (g *NaiveGreetingMaker) Hello(ctx context.Context, params greeting.Params) (string, error) {
	g.mu.RLock()
	val, found := g.data[params]
	g.mu.RUnlock()

	if !found {
		g.stats.Add(ctx, cache.MetricMiss, 1, "name", "greeting-naive")
	}

	expired := found && val.expires.Before(time.Now())
	if expired {
		g.stats.Add(ctx, cache.MetricExpired, 1, "name", "greeting-naive")
	}

	if !found || expired {
		g.stats.Add(ctx, cache.MetricWrite, 1, "name", "greeting-naive")

		gr, err := g.upstream.Hello(ctx, params)
		if err != nil {
			g.stats.Add(ctx, cache.MetricFailed, 1, "name", "greeting-naive")
			return gr, err
		}

		val.value = gr
		val.expires = time.Now().Add(g.ttl)

		g.mu.Lock()
		defer g.mu.Unlock()

		g.data[params] = val

		g.stats.Set(ctx, cache.MetricItems, float64(len(g.data)), "name", "greeting-naive")

		return val.value, nil
	}

	g.stats.Add(ctx, cache.MetricHit, 1, "name", "greeting-naive")

	return val.value, nil
}
