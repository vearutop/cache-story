package service

import (
	"github.com/bool64/brick"
)

// Locator defines application resources.
type Locator struct {
	*brick.BaseLocator

	GreetingMakerProvider
	GreetingClearerProvider
}
