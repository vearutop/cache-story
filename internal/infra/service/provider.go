package service

import (
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

// GreetingMakerProvider is a service provider.
type GreetingMakerProvider interface {
	GreetingMaker() greeting.Maker
}

// GreetingClearerProvider is a service provider.
type GreetingClearerProvider interface {
	GreetingClearer() greeting.Clearer
}
