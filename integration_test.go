package main_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/bool64/brick"
	"github.com/bool64/brick/config"
	"github.com/bool64/brick/test"
	"github.com/bool64/httptestbench"
	"github.com/godogx/dbsteps"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/vearutop/cache-story/internal/infra"
	"github.com/vearutop/cache-story/internal/infra/nethttp"
	"github.com/vearutop/cache-story/internal/infra/service"
	"github.com/vearutop/cache-story/internal/infra/storage"
)

func TestFeatures(t *testing.T) {
	var cfg service.Config

	test.RunFeatures(t, "", &cfg, func(tc *test.Context) (*brick.BaseLocator, http.Handler) {
		cfg.ServiceName = service.Name

		sl, err := infra.NewServiceLocator(cfg)
		require.NoError(t, err)

		tc.Database.Instances[dbsteps.Default] = dbsteps.Instance{
			Tables: map[string]interface{}{
				storage.GreetingsTable: new(storage.GreetingRow),
			},
		}

		return sl.BaseLocator, nethttp.NewRouter(sl)
	})
}

// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// cache: none
// BenchmarkGreetings-12    	     670	   1860402 ns/op	        85.60 50%:ms	       158.7 90%:ms	       238.0 99%:ms	       260.4 99.9%:ms	       151.8 B:rcvd/op	        92.82 B:sent/op	       537.4 rps	   54264 B/op	     850 allocs/op
// cache: naive
// BenchmarkGreetings-12    	     943	   1138846 ns/op	        49.64 50%:ms	       123.3 90%:ms	       241.4 99%:ms	       301.5 99.9%:ms	       151.8 B:rcvd/op	        92.79 B:sent/op	       878.0 rps	   44638 B/op	     684 allocs/op
// cache: advanced
// BenchmarkGreetings-12    	   41703	     25308 ns/op	         0.6497 50%:ms	         1.997 90%:ms	        13.42 99%:ms	        28.23 99.9%:ms	       151.9 B:rcvd/op	        92.90 B:sent/op	     39505 rps	   17090 B/op	     236 allocs/op
func BenchmarkGreetings(b *testing.B) {
	var cfg service.Config
	cfg.ServiceName = service.Name

	require.NoError(b, config.Load("", &cfg, config.WithOptionalEnvFiles(".env.integration-test")))

	sl, err := infra.NewServiceLocator(cfg)
	if err != nil {
		b.Skip(err)
	}

	router := nethttp.NewRouter(sl)

	srv := httptest.NewServer(router)
	defer srv.Close()

	httptestbench.RoundTrip(b, 50,
		func(i int, req *fasthttp.Request) {
			req.SetRequestURI(srv.URL + "/hello?locale=en-US&name=user" + strconv.Itoa(((i/10)^12345)%100))
		},
		func(i int, resp *fasthttp.Response) bool {
			return resp.StatusCode() == http.StatusOK
		},
	)
}
