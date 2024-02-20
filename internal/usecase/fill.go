package usecase

import (
	"context"
	"strconv"

	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

// Fill removes all saved greetings.
func Fill(deps interface {
	GreetingMaker() greeting.Maker
},
) usecase.Interactor {
	type input struct {
		N int `json:"n"`
	}

	type clearOutput struct {
		Affected int `json:"affected"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, in input, out *clearOutput) error {
		gm := deps.GreetingMaker()

		for i := 0; i < in.N; i++ {
			_, _ = gm.Hello(ctx, greeting.Params{
				Name:   "looooooooooooooooooooooooooooongstring" + strconv.Itoa(i),
				Locale: "en-US",
			})
		}

		return nil
	})

	u.SetDescription("Fill removes all saved greetings.")
	u.SetTags("Greeting")
	u.SetExpectedErrors(status.Unknown)

	return u
}
