// Package cached provides caching layer for domain services.
package cached

import (
	"context"

	"github.com/bool64/cache"
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

// NewGreetingMaker creates an instance of cached greeting maker.
func NewGreetingMaker(upstream greeting.Maker, cache *cache.FailoverOf[string]) *GreetingMaker {
	return &GreetingMaker{
		upstream: upstream,
		cache:    cache,
	}
}

// GreetingMaker uses cached value if available of fallbacks to upstream.
type GreetingMaker struct {
	upstream greeting.Maker
	cache    *cache.FailoverOf[string]
}

// GreetingMaker is a service provider.
func (g *GreetingMaker) GreetingMaker() greeting.Maker {
	return g
}

// Hello serves greeting.
func (g *GreetingMaker) Hello(ctx context.Context, params greeting.Params) (string, error) {
	key := []byte(params.Name + params.Locale)

	return g.cache.Get(ctx, key, func(ctx context.Context) (string, error) {
		return g.upstream.Hello(ctx, params)
	})
}
