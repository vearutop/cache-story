package main_test

import (
	"testing"

	"github.com/bool64/brick-template/internal/infra"
	"github.com/bool64/brick-template/internal/infra/nethttp"
	"github.com/bool64/brick-template/internal/infra/service"
	"github.com/bool64/brick-template/internal/infra/storage"
	"github.com/bool64/brick/config"
	"github.com/bool64/dbdog"
	"github.com/bool64/httpdog"
	"github.com/bool64/shared"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/require"
)

func TestFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	var (
		vars = &shared.Vars{}
		cfg  service.Config
	)

	require.NoError(t, config.Load("", &cfg, config.WithOptionalEnvFiles(".env.integration-test")))

	cfg.ServiceName = service.Name

	sl, err := infra.NewServiceLocator(cfg)
	require.NoError(t, err)

	addr, err := sl.StartHTTPServer(nethttp.NewRouter(sl))
	require.NoError(t, err)

	local := httpdog.NewLocal("http://" + addr)
	local.JSONComparer.Vars = vars

	dbm := dbdog.NewManager()
	dbm.Vars = vars
	dbm.Instances[dbdog.DefaultDatabase] = dbdog.Instance{
		Storage: sl.Storage,
		Tables: map[string]interface{}{
			storage.GreetingsTable: new(storage.GreetingRow),
		},
	}

	suite := godog.TestSuite{
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			local.RegisterSteps(s)
			dbm.RegisterSteps(s)
		},
		Options: &godog.Options{
			Format:      "pretty",
			Strict:      true,
			Concurrency: 1,
			Paths:       []string{"features"},
		},
	}

	status := suite.Run()

	sl.Shutdown()
	<-sl.Wait()

	if status != 0 {
		t.Fatal("test failed")
	}
}
