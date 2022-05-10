package cached

import (
	"context"
	"sync"
	"time"

	"github.com/bool64/cache"
	"github.com/bool64/stats"
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

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
		g.stats.Add(ctx, cache.MetricMiss, 1, "name", "greetings-naive")
	}

	expired := found && val.expires.Before(time.Now())
	if expired {
		g.stats.Add(ctx, cache.MetricExpired, 1, "name", "greetings-naive")
	}

	if !found || expired {
		g.stats.Add(ctx, cache.MetricWrite, 1, "name", "greetings-naive")

		gr, err := g.upstream.Hello(ctx, params)
		if err != nil {
			g.stats.Add(ctx, cache.MetricFailed, 1, "name", "greetings-naive")
			return gr, err
		}

		val.value = gr
		val.expires = time.Now().Add(g.ttl)

		g.mu.Lock()
		defer g.mu.Unlock()

		g.data[params] = val

		g.stats.Set(ctx, cache.MetricItems, float64(len(g.data)), "name", "greetings-naive")

		return val.value, nil
	}

	g.stats.Add(ctx, cache.MetricHit, 1, "name", "greetings-naive")

	return val.value, nil
}
