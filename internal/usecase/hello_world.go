package usecase

import (
	"context"

	"github.com/bool64/ctxd"
	"github.com/bool64/stats"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

type helloDeps interface {
	CtxdLogger() ctxd.Logger
	StatsTracker() stats.Tracker
	GreetingMaker() greeting.Maker
}

// HelloWorld creates use case interactor.
func HelloWorld(deps helloDeps) usecase.Interactor {
	type helloOutput struct {
		Message string `json:"message"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, in greeting.Params, out *helloOutput) error {
		deps.StatsTracker().Add(ctx, "hello", 1)
		deps.CtxdLogger().Info(ctx, "hello", "name", in.Name)

		msg, err := deps.GreetingMaker().Hello(ctx, in)

		out.Message = msg

		return err
	})

	u.SetDescription("Greeter says hello.")
	u.SetTags("Greeting")
	u.SetExpectedErrors(status.Unknown, status.InvalidArgument)

	return u
}
