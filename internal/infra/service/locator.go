package service

import (
	"github.com/bool64/brick"
	"github.com/bool64/sqluct"
)

// Locator defines application resources.
type Locator struct {
	*brick.BaseLocator

	Storage *sqluct.Storage

	GreetingMakerProvider
	GreetingClearerProvider
}
