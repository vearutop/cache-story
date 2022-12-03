package service

import (
	"github.com/bool64/brick"
	"github.com/bool64/brick/database"
	"github.com/bool64/brick/jaeger"
)

// Name is the name of this application or service.
const Name = "cache-story"

// Config defines application configuration.
type Config struct {
	brick.BaseConfig

	Cache string `split_words:"true" default:"advanced" enum:"none,naive,advanced,arena"`

	Database database.Config `split_words:"true"`
	Jaeger   jaeger.Config   `split_words:"true"`
}
