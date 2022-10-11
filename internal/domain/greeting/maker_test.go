package greeting_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

func TestSimpleMaker_GreetingMaker(t *testing.T) {
	sm := &greeting.SimpleMaker{}

	assert.Equal(t, sm, sm.GreetingMaker())
}
