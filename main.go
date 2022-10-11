// Package main provides cache-story web service.
package main

import (
	"log"
	"net/http"

	"github.com/bool64/brick"
	"github.com/vearutop/cache-story/internal/infra"
	"github.com/vearutop/cache-story/internal/infra/nethttp"
	"github.com/vearutop/cache-story/internal/infra/service"
)

func main() {
	var cfg service.Config

	brick.Start(&cfg, func(docsMode bool) (*brick.BaseLocator, http.Handler) {
		// Initialize application resources.
		sl, err := infra.NewServiceLocator(cfg)
		if err != nil {
			log.Fatalf("failed to init service: %v", err)
		}

		return sl.BaseLocator, nethttp.NewRouter(sl)
	})
}
