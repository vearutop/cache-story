// Package greeting defines greeting domain.
package greeting

import (
	"context"
	"strings"

	"github.com/bool64/ctxd"
)

// Params describes greeting input.
type Params struct {
	Name   string `query:"name" default:"World"`
	Locale string `query:"locale" required:"true" enum:"en-US,ru-RU"`
}

// Maker makes a greeting.
type Maker interface {
	Hello(ctx context.Context, params Params) (string, error)
}

// Clearer removes all greetings and returns number of affected rows.
type Clearer interface {
	ClearGreetings(ctx context.Context) (int, error)
}

// SimpleMaker can greet you in two locales.
type SimpleMaker struct{}

// Hello greets.
func (s *SimpleMaker) Hello(ctx context.Context, params Params) (string, error) {
	if strings.ToLower(params.Name) == "bug" {
		return "", ctxd.NewError(ctx, "#$@@^! %C 🤖")
	}

	switch params.Locale {
	case "en-US":
		return "Hello, " + params.Name + "!", nil
	case "ru-RU":
		return "Привет, " + params.Name + "!", nil
	default:
		return "", ctxd.NewError(ctx, "unknown locale", "locale", params.Locale)
	}
}

// GreetingMaker implements service provider.
func (s *SimpleMaker) GreetingMaker() Maker {
	if s == nil {
		panic("empty SimpleMaker")
	}

	return s
}
