package schema

import (
	"github.com/swaggest/rest/openapi"
)

// SetupOpenapiCollector configures OpenAPI schema.
func SetupOpenapiCollector(c *openapi.Collector) {
	c.Reflector().SpecEns().Info.Title = "brick-template"
}
