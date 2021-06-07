// Package ui provides application web user interface.
package ui

import (
	"net/http"
	"os"

	"github.com/bool64/brick-template/resources/static"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

// Static serves static assets.
var Static http.Handler

// nolint:gochecknoinits
func init() {
	if _, err := os.Stat("./resources/static"); err == nil {
		// path/to/whatever exists
		Static = http.FileServer(http.Dir("./resources/static"))
	} else {
		Static = statigz.FileServer(static.Assets, brotli.AddEncoding, statigz.EncodeOnInit)
	}
}

// Index serves index page of the application.
func Index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Static.ServeHTTP(w, r)
	})
}
