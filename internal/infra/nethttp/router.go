// Package nethttp manages application http interface.
package nethttp

import (
	"net/http"

	"github.com/bool64/brick"
	"github.com/swaggest/rest/nethttp"
	"github.com/vearutop/cache-story/internal/infra/nethttp/ui"
	"github.com/vearutop/cache-story/internal/infra/service"
	"github.com/vearutop/cache-story/internal/usecase"
)

// NewRouter creates an instance of router filled with handlers and docs.
func NewRouter(deps *service.Locator) http.Handler {
	r := brick.NewBaseRouter(deps.BaseLocator)

	r.Method(http.MethodGet, "/hello", nethttp.NewHandler(usecase.HelloWorld(deps)))

	r.Method(http.MethodGet, "/", ui.Index())
	r.Mount("/static/", http.StripPrefix("/static", ui.Static))

	return r
}
