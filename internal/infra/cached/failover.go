package cached

import (
	"context"

	"github.com/bool64/cache"
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
