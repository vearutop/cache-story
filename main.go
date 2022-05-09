package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bool64/brick"
	"github.com/bool64/brick/config"
	"github.com/bool64/dev/version"
	"github.com/swaggest/assertjson"
	"github.com/vearutop/cache-story/internal/infra"
	"github.com/vearutop/cache-story/internal/infra/nethttp"
	"github.com/vearutop/cache-story/internal/infra/schema"
	"github.com/vearutop/cache-story/internal/infra/service"
)

func main() {
	ver := flag.Bool("version", false, "Print application version and exit.")
	api := flag.Bool("openapi", false, "Print application OpenAPI spec and exit.")
	flag.Parse()

	if ver != nil && *ver {
		fmt.Println(version.Info().Version)

		return
	}

	if api != nil && *api {
		printOpenAPI()

		return
	}

	var cfg service.Config

	if err := config.Load("", &cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize application resources.
	sl, err := infra.NewServiceLocator(cfg)
	if err != nil {
		log.Fatalf("failed to init service: %v", err)
	}

	_, err = sl.StartHTTPServer(nethttp.NewRouter(sl))
	if err != nil {
		sl.CtxdLogger().Error(context.Background(), "failed to start http server: %v", "error", err)
		os.Exit(1)
	}

	// Wait for service sl termination finished.
	err = <-sl.Wait()
	if err != nil {
		sl.CtxdLogger().Error(context.Background(), err.Error())
	}
}

func printOpenAPI() {
	l := &service.Locator{BaseLocator: brick.NoOpLocator()}
	schema.SetupOpenapiCollector(l.OpenAPI)
	nethttp.NewRouter(l)

	j, err := assertjson.MarshalIndentCompact(l.OpenAPI.Reflector().Spec, "", " ", 80)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(j))
}
