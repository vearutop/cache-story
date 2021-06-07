package service

import (
	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/bool64/brick"
	"github.com/bool64/brick/database"
)

// Name is the name of this application or service.
const Name = "brick-template"

// Config defines application configuration.
type Config struct {
	brick.BaseConfig

	Database database.Config `split_words:"true"`
	Jaeger   jaeger.Options  `split_words:"true"`
}
