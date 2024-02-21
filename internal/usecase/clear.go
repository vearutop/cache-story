package usecase

import (
	"context"

	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

// Clear removes all saved greetings.
func Clear(deps interface {
	GreetingClearer() greeting.Clearer
},
) usecase.Interactor {
	type clearOutput struct {
		Affected int `json:"affected"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, out *clearOutput) error {
		affected, err := deps.GreetingClearer().ClearGreetings(ctx)

		out.Affected = affected

		return err
	})

	u.SetDescription("Clear removes all saved greetings.")
	u.SetTags("Greeting")
	u.SetExpectedErrors(status.Unknown)

	return u
}
